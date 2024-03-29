package http

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"example/grpc/pb"
)

type handler struct {
	client pb.TestClient
	db     *sql.DB
}

func New(client pb.TestClient, db *sql.DB) *handler {
	return &handler{
		client: client,
		db:     db,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n1, err := strconv.ParseInt(r.FormValue("n1"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	n2, err := strconv.ParseInt(r.FormValue("n2"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	resp, err := h.client.Sum(ctx, &pb.SumReq{
		N1: n1,
		N2: n2,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(w, "Sum = %d", resp.Sum)
}
