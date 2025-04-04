package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/TechBowl-japan/go-stations/model"
	"github.com/TechBowl-japan/go-stations/service"
)

// A TODOHandler implements handling REST endpoints.
type TODOHandler struct {
	svc *service.TODOService
}

// NewTODOHandler returns TODOHandler based http.Handler.
func NewTODOHandler(svc *service.TODOService) *TODOHandler {
	return &TODOHandler{
		svc: svc,
	}
}

func (h *TODOHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req model.CreateTODORequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad REQUEST", http.StatusBadRequest)
			return
		}
		if req.Subject == "" {
			http.Error(w, "subject is required", http.StatusBadRequest)
			return
		}
		res, err := h.Create(r.Context(), &req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Create error: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)

	case http.MethodPut:
		var req model.UpdateTODORequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if req.Subject == "" {
			http.Error(w, "subject is required", http.StatusBadRequest)
			return
		}
		if req.ID == 0 {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}
		res, err := h.Update(r.Context(), &req)
		if err != nil {
			var notFoundErr *model.ErrNotFound
			if errors.As(err, &notFoundErr) {
				http.Error(w, "TODO not found", http.StatusNotFound)
			} else {
				http.Error(w, "failed to update", http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)

	case http.MethodGet:
		// クエリパラメータ取得
		query := r.URL.Query()

		// prev_id を int64 に変換（空なら 0）
		prevIDStr := query.Get("prev_id")
		sizeStr := query.Get("size")

		var prevID, size int64
		var err error

		if prevIDStr != "" {
			prevID, err = strconv.ParseInt(prevIDStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid prev_id", http.StatusBadRequest)
				return
			}
		}
		if sizeStr != "" {
			size, err = strconv.ParseInt(sizeStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid size", http.StatusBadRequest)
				return
			}
		}

		if size == 0 {
			size = 5
		}

		// ReadTODO を呼び出す
		todos, err := h.svc.ReadTODO(r.Context(), prevID, size)
		if err != nil {
			http.Error(w, "failed to read todos", http.StatusInternalServerError)
			return
		}

		// レスポンス生成して返す
		res := &model.ReadTODOResponse{
			TODOs: todos,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)

	case http.MethodDelete:
		var req model.DeleteTODORequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if len(req.IDs) == 0 {
			http.Error(w, "IDs are required", http.StatusBadRequest)
			return
		}
		err := h.svc.DeleteTODO(r.Context(), req.IDs)
		if err != nil {
			var notFoundErr *model.ErrNotFound
			if errors.As(err, &notFoundErr) {
				http.Error(w, "Todo not found", http.StatusNotFound)
			} else {
				http.Error(w, "failed to delete", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(&model.DeleteTODOResponse{})

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// Create handles the endpoint that creates the TODO.
func (h *TODOHandler) Create(ctx context.Context, req *model.CreateTODORequest) (*model.CreateTODOResponse, error) {
	if req.Subject == "" {
		return nil, fmt.Errorf("subject is required")
	}
	todo, err := h.svc.CreateTODO(ctx, req.Subject, req.Description)
	if err != nil {
		return nil, err
	}

	return &model.CreateTODOResponse{
		TODO: *todo,
	}, nil
}

// Read handles the endpoint that reads the TODOs.
func (h *TODOHandler) Read(ctx context.Context, req *model.ReadTODORequest) (*model.ReadTODOResponse, error) {
	todos, err := h.svc.ReadTODO(ctx, req.PrevID, req.Size)
	if err != nil {
		return nil, err
	}

	return &model.ReadTODOResponse{
		TODOs: todos,
	}, nil
}

// Update handles the endpoint that updates the TODO.
func (h *TODOHandler) Update(ctx context.Context, req *model.UpdateTODORequest) (*model.UpdateTODOResponse, error) {
	todo, err := h.svc.UpdateTODO(ctx, req.ID, req.Subject, req.Description)
	if err != nil {
		return nil, err
	}

	return &model.UpdateTODOResponse{
		TODO: *todo,
	}, nil
}

// Delete handles the endpoint that deletes the TODOs.
func (h *TODOHandler) Delete(ctx context.Context, req *model.DeleteTODORequest) (*model.DeleteTODOResponse, error) {

	err := h.svc.DeleteTODO(ctx, req.IDs)
	if err != nil {
		return nil, err
	}

	return &model.DeleteTODOResponse{}, nil
}
