package http

import (
	"billing-api/internal/service"
	"net/http"
)

type Handler struct {
	billingService *service.BillingService
}

func NewHandler(bs *service.BillingService) *Handler {
	return &Handler{
		billingService: bs,
	}
}

func (h *Handler) GetOutstanding(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("api not implemented yet"))
}

func (h *Handler) MakePayment(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("api not implemented yet"))
}
