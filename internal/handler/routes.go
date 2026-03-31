package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/semmidev/atstex-lab/docs"
	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/middleware"
	"github.com/semmidev/atstex-lab/web"
)

func (s *Server) routes() {
	s.router = chi.NewRouter()

	// Middleware stack.
	s.router.Use(chimw.RequestID)
	s.router.Use(chimw.RealIP)
	s.router.Use(middleware.SecurityHeaders)
	s.router.Use(middleware.MaxBodySize(1 << 20)) // 1MB default body limit
	s.router.Use(middleware.RequestLogger(s.logger))
	s.router.Use(chimw.Recoverer)
	s.router.Use(chimw.Timeout(120 * time.Second))
	s.router.Use(chimw.CleanPath)
	s.router.Use(chimw.StripSlashes)

	// Swagger UI
	s.router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

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
		r.Get("/analytics", s.handleAnalyticsPage())
		r.Get("/subscription", s.handleSubscriptionPage())
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
		r.Put("/api/cv-profiles/{id}/title", s.handleUpdateCVProfileTitle())
		r.Delete("/api/cv-profiles/{id}", s.handleDeleteCVProfile())
		r.Put("/api/cv-profiles/{id}/visibility", s.handleToggleProfileVisibility())
		r.Post("/api/cv-profiles/{id}/auto-tailor", s.handleAutoTailorCVProfile())
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
		// ATS Simulator
		r.Get("/ats-simulator", s.handleAtsSimulatorPage())
		r.Post("/api/ats-simulator", s.handleCreateAtsSimulation())
		r.Get("/api/ats-simulations", s.handleListMyAtsSimulations())
		// AI Interview Prep
		r.Get("/interview-prep", s.handleInterviewPrepPage())
		r.Post("/api/interview-prep", s.handleCreateInterviewPrep())
		r.Get("/api/interview-preps", s.handleListMyInterviewPreps())
		r.Post("/api/interview-prep/critique", s.handleCritiqueInterviewAnswer())
		// Mock Interview (Live WebSockets)
		r.Get("/mock-interview", s.handleMockInterviewPage())
		r.Get("/api/mock-interview/sessions", s.handleListMockInterviewSessions())
		r.Get("/ws/mock-interview", s.handleMockInterviewWS())
		// AI Bullet Point Enhancer
		r.Post("/api/ai/enhance-bullet", s.handleEnhanceBullet())
		// Gallery (Multi-Template Preview)
		r.Get("/gallery", s.handleGalleryPage())
		r.Post("/api/gallery/compile", s.handleGalleryCompile())

		// Custom Template Builder
		r.Get("/builder", s.handleBuilderPage())
		r.Post("/api/builder/save", s.handleSaveCustomTemplate())
		r.Get("/api/builder/load", s.handleLoadCustomTemplates())
		r.Post("/api/builder/preview", s.handlePreviewCustomTemplate())
	})

	// Admin routes (requires login + admin role)
	s.router.Group(func(r chi.Router) {
		r.Use(auth.Middleware(s.repo))
		r.Use(auth.AdminMiddleware())
		r.Get("/admin", s.handleAdminDashboard())
		r.Get("/api/admin/stats", s.handleAdminGetStats())
		r.Get("/api/admin/analytics", s.handleAdminAnalytics())
		r.Get("/api/admin/users", s.handleAdminListUsers())
		r.Get("/api/admin/feedbacks", s.handleAdminListFeedbacks())
		r.Post("/api/admin/feedbacks/{id}/reply", s.handleAdminReplyFeedback())
		r.Delete("/api/admin/feedbacks/{id}", s.handleAdminDeleteFeedback())
		r.Post("/api/admin/users/{id}/block", s.handleAdminBlockUser())
		r.Post("/api/admin/users/{id}/unblock", s.handleAdminUnblockUser())
		r.Post("/api/admin/users/{id}/make-admin", s.handleAdminMakeUserAdmin())
		r.Post("/api/admin/users/{id}/revoke-admin", s.handleAdminRevokeUserAdmin())
		r.Delete("/api/admin/users/{id}", s.handleAdminDeleteUser())

		// Subscription Management
		r.Get("/api/admin/subscription-plans", s.handleAdminListSubscriptionPlans())
		r.Post("/api/admin/subscription-plans", s.handleAdminCreateSubscriptionPlan())
		r.Put("/api/admin/subscription-plans/{id}", s.handleAdminUpdateSubscriptionPlan())
		r.Delete("/api/admin/subscription-plans/{id}", s.handleAdminDeleteSubscriptionPlan())
		r.Post("/api/admin/subscription-plans/{id}/toggle", s.handleAdminToggleSubscriptionPlan())
		r.Post("/api/admin/users/{id}/subscribe", s.handleAdminAssignSubscription())
		r.Get("/api/admin/users/{id}/subscription", s.handleAdminGetUserSubscription())
	})

	// Forbidden page (accessible without login)
	s.router.Get("/forbidden", s.handleForbiddenPage())

	// Serve static assets (JS/CSS) embedded in the binary.
	s.router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(web.StaticFS))))
}
