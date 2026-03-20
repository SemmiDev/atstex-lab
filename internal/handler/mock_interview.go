package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/semmidev/atstex-lab/internal/aisuites"
	"github.com/semmidev/atstex-lab/internal/apperrors"
	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/semmidev/atstex-lab/internal/middleware"
)

// handleMockInterviewPage renders the Mock Interview setup + room UI.
func (s *Server) handleMockInterviewPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		profiles, _ := s.repo.GetCVProfilesByUserID(r.Context(), user.ID)

		if err := s.tmpl.ExecuteTemplate(w, "mock-interview", map[string]interface{}{
			"User":     user,
			"Profiles": profiles,
		}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// handleListMockInterviewSessions returns the current user's past sessions (JSON).
func (s *Server) handleListMockInterviewSessions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		sessions, err := s.repo.GetMockInterviewSessionsByUserID(r.Context(), user.ID)
		if err != nil {
			s.reqLog(r).Error("list mock interview sessions error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to list sessions")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, sessions)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// WebSocket handler — the core of the live mock interview
// ─────────────────────────────────────────────────────────────────────────────

// handleMockInterviewWS upgrades the connection to WebSocket and manages a
// full conversational AI mock interview session.
//
// Query parameters (sent on the initial HTTP request before the upgrade):
//
//	profileId      – CV profile UUID
//	language       – interview language code (en / id / ja / zh / ko)
//	jobDescription – base64-encoded job description (to avoid URL length limits)
//	interviewerStyle – interviewer profile/style (e.g. balanced, friendly, strict, technical, behavioral)
func (s *Server) handleMockInterviewWS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		// ── 1. Validate query parameters ────────────────────────────────────
		profileIDStr := strings.TrimSpace(r.URL.Query().Get("profileId"))
		language := strings.TrimSpace(r.URL.Query().Get("language"))
		jobDescription := strings.TrimSpace(r.URL.Query().Get("jobDescription"))
		interviewerStyle := strings.TrimSpace(r.URL.Query().Get("interviewerStyle"))

		if profileIDStr == "" {
			http.Error(w, "profileId is required", http.StatusBadRequest)
			return
		}

		profileID, err := uuid.Parse(profileIDStr)
		if err != nil {
			http.Error(w, "invalid profileId", http.StatusBadRequest)
			return
		}

		if language == "" {
			language = "en"
		}

		if interviewerStyle == "" {
			interviewerStyle = "balanced"
		}
		switch strings.ToLower(interviewerStyle) {
		case "balanced", "friendly", "strict", "technical", "behavioral":
			// ok
		default:
			http.Error(w, "invalid interviewerStyle", http.StatusBadRequest)
			return
		}

		// ── 2. Load CV profile ───────────────────────────────────────────────
		profile, err := s.repo.GetCVProfile(r.Context(), profileID)
		if err != nil {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}
		if profile.UserID != user.ID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if len(profile.Biodata) == 0 || string(profile.Biodata) == "null" || string(profile.Biodata) == "{}" {
			http.Error(w, "CV profile has no biodata — please fill in your biodata first", http.StatusBadRequest)
			return
		}

		// ── 3. Upgrade to WebSocket ──────────────────────────────────────────
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true, // handled by auth middleware upstream
		})
		if err != nil {
			s.reqLog(r).Error("websocket accept error", "err", err)
			return
		}
		defer conn.Close(websocket.StatusInternalError, "connection closed")

		// Use a generous timeout for the whole session lifetime.
		ctx, cancel := context.WithTimeout(r.Context(), 90*time.Minute)
		defer cancel()

		// ── 4. Create a DB session record ────────────────────────────────────
		session := &domain.MockInterviewSession{
			UserID:           user.ID,
			ProfileID:        profileID,
			JobDescription:   jobDescription,
			Language:         language,
			InterviewerStyle: strings.ToLower(interviewerStyle),
			Messages:         json.RawMessage("[]"),
		}
		if dbErr := s.repo.CreateMockInterviewSession(ctx, session); dbErr != nil {
			s.reqLog(r).Error("create mock interview session error", "err", dbErr)
			_ = wsjson.Write(ctx, conn, domain.MockInterviewWSMessage{
				Type:  domain.WSTypeError,
				Error: "failed to create session",
			})
			return
		}

		// ── 5. Run the conversation loop ─────────────────────────────────────
		s.runMockInterviewLoop(ctx, conn, session, string(profile.Biodata), user)
	}
}

// runMockInterviewLoop is the main conversation engine.
// It sends the first AI greeting, then alternates: user speaks → AI responds.
func (s *Server) runMockInterviewLoop(
	ctx context.Context,
	conn *websocket.Conn,
	session *domain.MockInterviewSession,
	biodataJSON string,
	user *domain.User,
) {
	log := s.logger.With("sessionId", session.ID, "userId", user.ID)

	// Conversation history kept entirely in memory during the session.
	// The aisuites package uses its own message type so we bridge here.
	var aiHistory []aisuites.MockInterviewMessage
	var domainMessages []domain.MockInterviewMessage
	var totalTokens int64

	// helper: persist current state to DB
	persistSession := func() {
		msgsJSON, err := json.Marshal(domainMessages)
		if err != nil {
			log.Error("marshal messages error", "err", err)
			return
		}
		session.Messages = json.RawMessage(msgsJSON)
		session.TurnCount = len(domainMessages)
		session.TokensUsed = totalTokens
		if err := s.repo.UpdateMockInterviewSession(ctx, session); err != nil {
			log.Error("persist session error", "err", err)
		}
	}

	// helper: send an AI message to the client
	sendAI := func(text string, turnID int) bool {
		msg := domain.MockInterviewWSMessage{
			Type:      domain.WSTypeAIMessage,
			Text:      text,
			SessionID: session.ID.String(),
			TurnID:    turnID,
		}
		if err := wsjson.Write(ctx, conn, msg); err != nil {
			log.Error("write ai message error", "err", err)
			return false
		}
		return true
	}

	// helper: send a thinking indicator
	sendThinking := func() {
		_ = wsjson.Write(ctx, conn, domain.MockInterviewWSMessage{
			Type:      domain.WSTypeThinking,
			SessionID: session.ID.String(),
		})
	}

	// ── Turn 0: AI starts the interview ─────────────────────────────────────
	sendThinking()

	aiText, newHistory, tokens, err := aisuites.HandleMockInterviewTurn(
		ctx,
		aiHistory, // empty → triggers system prompt + first greeting
		biodataJSON,
		session.JobDescription,
		session.Language,
		session.InterviewerStyle,
		s.aiConfig,
	)
	if err != nil {
		log.Error("first AI turn error", "err", err)
		_ = wsjson.Write(ctx, conn, domain.MockInterviewWSMessage{
			Type:  domain.WSTypeError,
			Error: "AI failed to start interview: " + err.Error(),
		})
		return
	}

	totalTokens += tokens
	aiHistory = newHistory

	// Record AI opening in domain messages (skip the system prompt)
	domainMessages = append(domainMessages, domain.MockInterviewMessage{
		Role:      "ai",
		Text:      aiText,
		CreatedAt: time.Now(),
	})

	if !sendAI(aiText, 0) {
		return
	}
	persistSession()

	turnID := 1

	// ── Conversation loop ────────────────────────────────────────────────────
	for {
		// Wait for a message from the client
		var incoming domain.MockInterviewWSMessage
		if errRead := wsjson.Read(ctx, conn, &incoming); errRead != nil {
			// Connection closed by client or timeout — end the session cleanly
			log.Info("websocket read ended", "err", err)
			break
		}

		switch incoming.Type {
		case domain.WSTypePing:
			_ = wsjson.Write(ctx, conn, domain.MockInterviewWSMessage{Type: domain.WSTypePong})
			continue

		case domain.WSTypeEndSession:
			// Client explicitly ends the session
			now := time.Now()
			session.EndedAt = &now
			persistSession()
			_ = wsjson.Write(ctx, conn, domain.MockInterviewWSMessage{
				Type:      domain.WSTypeEnded,
				SessionID: session.ID.String(),
				TurnID:    turnID,
			})
			log.Info("mock interview session ended by user", "turns", turnID, "tokens", totalTokens)
			conn.Close(websocket.StatusNormalClosure, "session ended")
			return

		case domain.WSTypeUserMsg:
			userText := strings.TrimSpace(incoming.Text)
			if userText == "" {
				continue
			}

			// Record user message
			domainMessages = append(domainMessages, domain.MockInterviewMessage{
				Role:      "user",
				Text:      userText,
				CreatedAt: time.Now(),
			})

			// Append to AI history as "human" role
			aiHistory = append(aiHistory, aisuites.MockInterviewMessage{
				Role:    "human",
				Content: userText,
			})

			// Ask AI for the next response
			sendThinking()

			aiText, newHistory, tokens, err = aisuites.HandleMockInterviewTurn(
				ctx,
				aiHistory,
				biodataJSON,
				session.JobDescription,
				session.Language,
				session.InterviewerStyle,
				s.aiConfig,
			)
			if err != nil {
				log.Error("AI turn error", "err", err, "turn", turnID)
				_ = wsjson.Write(ctx, conn, domain.MockInterviewWSMessage{
					Type:  domain.WSTypeError,
					Error: "AI response failed: " + err.Error(),
				})
				// Don't break — allow the client to retry
				continue
			}

			totalTokens += tokens
			aiHistory = newHistory

			// Increment AI tokens in the user's usage tracker (best-effort)
			if tokens > 0 {
				_ = s.repo.IncrementAITokensUsed(ctx, user.ID, tokens)
			}

			// Record AI response
			domainMessages = append(domainMessages, domain.MockInterviewMessage{
				Role:      "ai",
				Text:      aiText,
				CreatedAt: time.Now(),
			})

			if !sendAI(aiText, turnID) {
				break
			}

			turnID++
			persistSession()

		default:
			// Unknown message type — ignore
			log.Warn("unknown ws message type", "type", incoming.Type)
		}
	}

	// ── Session ended (connection dropped) ───────────────────────────────────
	now := time.Now()
	session.EndedAt = &now
	persistSession()
	log.Info("mock interview session closed", "turns", turnID, "tokens", totalTokens)
}
