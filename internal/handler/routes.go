package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/semmidev/atstex-lab/internal/auth"
	mw "github.com/semmidev/atstex-lab/internal/middleware"
	"github.com/semmidev/atstex-lab/web"
)

func (s *Server) routes() {
	s.router = chi.NewRouter()

	// Middleware stack.
	s.router.Use(chimw.RequestID)
	s.router.Use(chimw.RealIP)
	s.router.Use(mw.RequestLogger(s.logger))
	s.router.Use(chimw.Recoverer)
	s.router.Use(chimw.Timeout(120 * time.Second))
	s.router.Use(chimw.CleanPath)
	s.router.Use(chimw.StripSlashes)

	// Public routes
	s.router.Group(func(r chi.Router) {
		r.Use(auth.OptionalUserMiddleware(s.repo))
		r.Get("/", s.handleHome())
		r.Get("/u/{username}", s.handlePublicProfile())
		r.Get("/u/{username}/{profileID}/pdf", s.handlePublicProfileDownloadPDF())
	})

	// Auth routes
	s.router.Get("/auth/google/login", s.handleGoogleLogin())
	s.router.Get("/auth/google/callback", s.handleGoogleCallback())
	s.router.Post("/auth/logout", s.handleLogout())
	s.router.Post("/auth/delete-account", s.handleDeleteAccount())

	// Protected routes
	s.router.Group(func(r chi.Router) {
		r.Use(auth.Middleware(s.repo))
		r.Get("/profile", s.handleProfile())
		r.Get("/support", s.handleSupport())
		r.Post("/auth/sessions/{token}/delete", s.handleDeleteSession())
		r.Get("/input", s.handleInput())
		r.Get("/input/embed", s.handleInputEmbed())
		r.Get("/editor", s.handleEditor())
		r.Get("/publish", s.handlePublishSettings())
		r.Get("/api/templates", s.handleListTemplates())
		r.Get("/api/templates/{name}", s.handleGetTemplate())
		r.Post("/api/templates/{name}/render", s.handleRenderTemplate())
		r.Post("/compile", s.handleCompile())
		r.Post("/api/extract-pdf", s.handleExtractPDF())
		r.Post("/api/page-settings/apply", s.handleApplyPageSettings())
		// CV Profile API
		r.Get("/api/cv-profiles", s.handleListCVProfiles())
		r.Post("/api/cv-profiles", s.handleCreateCVProfile())
		r.Get("/api/cv-profiles/{id}", s.handleGetCVProfile())
		r.Put("/api/cv-profiles/{id}", s.handleSaveCVProfile())
		r.Delete("/api/cv-profiles/{id}", s.handleDeleteCVProfile())
		r.Put("/api/cv-profiles/{id}/visibility", s.handleToggleProfileVisibility())
		// Public profile API
		r.Put("/api/username", s.handleSetUsername())
		r.Get("/api/username/check", s.handleCheckUsername())
		// Feedback
		r.Get("/feedback", s.handleFeedbackPage())
		r.Get("/api/feedback", s.handleListMyFeedbacks())
		r.Post("/api/feedback", s.handleCreateFeedback())
		// CV Review
		r.Get("/cv-review", s.handleCVReviewPage())
		r.Post("/api/cv-review", s.handleCreateCVReview())
		r.Get("/api/cv-reviews", s.handleListMyCVReviews())
		// Cover Letter
		r.Get("/cover-letter", s.handleCoverLetterPage())
		r.Post("/api/cover-letter/generate", s.handleGenerateCoverLetter())
		r.Get("/api/cover-letters", s.handleListMyCoverLetters())
		// Job Application Tracking
		r.Get("/kanban", s.handleKanbanPage())
		r.Get("/api/applications", s.GetJobApplications())
		r.Post("/api/applications", s.CreateJobApplication())
		r.Put("/api/applications/{id}", s.UpdateJobApplication())
		r.Put("/api/applications/{id}/status", s.UpdateJobApplicationStatus())
		r.Delete("/api/applications/{id}", s.DeleteJobApplication())
	})

	// Admin routes (requires login + admin role)
	s.router.Group(func(r chi.Router) {
		r.Use(auth.Middleware(s.repo))
		r.Use(auth.AdminMiddleware())
		r.Get("/admin", s.handleAdminDashboard())
		r.Get("/api/admin/stats", s.handleAdminGetStats())
		r.Get("/api/admin/users", s.handleAdminListUsers())
		r.Get("/api/admin/feedbacks", s.handleAdminListFeedbacks())
		r.Post("/api/admin/feedbacks/{id}/reply", s.handleAdminReplyFeedback())
		r.Delete("/api/admin/feedbacks/{id}", s.handleAdminDeleteFeedback())
		r.Post("/api/admin/users/{id}/block", s.handleAdminBlockUser())
		r.Post("/api/admin/users/{id}/unblock", s.handleAdminUnblockUser())
		r.Post("/api/admin/users/{id}/make-admin", s.handleAdminMakeUserAdmin())
		r.Post("/api/admin/users/{id}/revoke-admin", s.handleAdminRevokeUserAdmin())
		r.Delete("/api/admin/users/{id}", s.handleAdminDeleteUser())
	})

	// Forbidden page (accessible without login)
	s.router.Get("/forbidden", s.handleForbiddenPage())

	// Serve static assets (JS/CSS) embedded in the binary.
	s.router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(web.StaticFS))))
}
