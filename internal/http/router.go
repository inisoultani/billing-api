package http

import (
	"billing-api/internal/config"
	"billing-api/internal/service"
	"net/http"

	billingApiMiddleware "billing-api/internal/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(billingService *service.BillingService, cfg *config.Config) http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Route("/loan", func(r chi.Router) {
		h := NewHandler(billingService, cfg)

		r.Post("/", h.SubmitLoan)
		r.Get("/{loanID}", h.MakeHandler(h.GetLoanByID))
		r.Get("/{loanID}/outstanding", h.GetOutstanding)
		r.Get("/{loanID}/payment", h.ListPayments)
		r.Get("/{loanID}/schedule", h.ListSchedules)

		r.Group(func(r chi.Router) {
			r.Use(billingApiMiddleware.IdempotencyMiddleware)
			r.Post("/{loanID}/payment", h.MakeHandler(h.MakePayment))
		})
	})

	return r

}
