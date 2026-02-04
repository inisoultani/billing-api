package http

import (
	"billing-api/internal/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(billingService *service.BillingService) http.Handler {

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
		h := NewHandler(billingService)
		r.Get("/{loanID}/outstanding", h.GetOutstanding)
		r.Post("/{loanID}/payment", h.MakePayment)
	})

	return r

}
