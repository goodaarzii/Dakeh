package users

import (
	"log"
	"net/http"

	"github.com/goodaarzii/dakeh/internal/json"
	"github.com/goodaarzii/dakeh/internal/usecase/user"
)

type Handler struct {
	srv user.Service
}

func NewHandler(service user.Service) *Handler {
	return &Handler{srv: service}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	products, err := h.srv.CreateUser(r.Context())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, products)
}
