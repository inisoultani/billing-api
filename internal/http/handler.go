package http

import "net/http"

func GetOutstanding(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("api not implemented yet"))
}

func MakePayment(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("api not implemented yet"))
}
