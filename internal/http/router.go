package http

import (
	"billing-api/internal/config"
	"billing-api/internal/service"
	"net/http"

	"billing-api/internal/http/handler"
	billingApiMiddleware "billing-api/internal/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(billingService *service.BillingService, cfg *config.Config) http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	// r.Use(billingApiMiddleware.RequestIDMiddleware)
	r.Use(middleware.RequestID)
	r.Use(billingApiMiddleware.LoggerMiddleware)
	r.Use(middleware.RealIP)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Route("/loan", func(r chi.Router) {
		h := handler.NewHandler(billingService, cfg)

		r.Post("/", h.MakeHandler(h.SubmitLoan))
		r.Get("/{loanID}", h.MakeHandler(h.GetLoanByID))
		r.Get("/{loanID}/outstanding", h.MakeHandler(h.GetOutstanding))
		r.Get("/{loanID}/payment", h.MakeHandler(h.ListPayments))
		r.Get("/{loanID}/schedule", h.MakeHandler(h.ListSchedules))

		r.Group(func(r chi.Router) {
			r.Use(billingApiMiddleware.IdempotencyMiddleware)
			r.Post("/{loanID}/payment", h.MakeHandler(h.MakePayment))
		})
		r.Group(func(r chi.Router) {
			// later we can put specific auth middleware here
			r.Post("/admin/log-level", h.MakeHandler(h.ChangeLogLevel(cfg.LogLevel)))
		})
	})

	return r

}
