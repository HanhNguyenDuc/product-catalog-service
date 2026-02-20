package rest

import (
	"encoding/json"
	"net/http"
	"time"

	activateproduct "github.com/product-catalog-service/internal/app/product/usecases/activate_product"
	applydiscount "github.com/product-catalog-service/internal/app/product/usecases/apply_discount"
	createproduct "github.com/product-catalog-service/internal/app/product/usecases/create_product"
	updateproduct "github.com/product-catalog-service/internal/app/product/usecases/update_product"
)

// ── Create ────────────────────────────────────────────────────────────────────

type createProductBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

func (s *Server) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var body createProductBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := s.p.CreateProductInteractor.Execute(r.Context(), &createproduct.CreateProductRequest{
		Name:        body.Name,
		Description: body.Description,
		Category:    body.Category,
	})
	if err != nil {
		s.p.Log.Sugar().Errorw("createProduct", "error", err)
		writeError(w, domainErrToStatus(err), err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// ── Update ────────────────────────────────────────────────────────────────────

type updateProductBody struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
}

func (s *Server) handleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var body updateProductBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := s.p.UpdateProductInteractor.Execute(r.Context(), &updateproduct.UpdateProductRequest{
		ProductID:   id,
		Name:        body.Name,
		Description: body.Description,
		Category:    body.Category,
	})
	if err != nil {
		s.p.Log.Sugar().Errorw("updateProduct", "id", id, "error", err)
		writeError(w, domainErrToStatus(err), err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Activate ─────────────────────────────────────────────────────────────────

func (s *Server) handleActivateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := s.p.ActivateProductInteractor.Execute(r.Context(), &activateproduct.ActivateProductRequest{
		ProductID: id,
	})
	if err != nil {
		s.p.Log.Sugar().Errorw("activateProduct", "id", id, "error", err)
		writeError(w, domainErrToStatus(err), err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Apply Discount ────────────────────────────────────────────────────────────

type applyDiscountBody struct {
	Percentage string    `json:"percentage"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
}

func (s *Server) handleApplyDiscount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var body applyDiscountBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := s.p.ApplyDiscountInteractor.Execute(r.Context(), &applydiscount.ApplyDiscountRequest{
		ProductID:  id,
		Percentage: body.Percentage,
		StartsAt:   body.StartsAt,
		EndsAt:     body.EndsAt,
	})
	if err != nil {
		s.p.Log.Sugar().Errorw("applyDiscount", "id", id, "error", err)
		writeError(w, domainErrToStatus(err), err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
