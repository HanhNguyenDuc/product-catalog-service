package rest

import (
	"net/http"
	"strconv"

	getproduct "github.com/product-catalog-service/internal/app/product/queries/get_product"
	listproducts "github.com/product-catalog-service/internal/app/product/queries/list_products"
)

// ── Get by ID ─────────────────────────────────────────────────────────────────

func (s *Server) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	dto, err := s.p.GetProductQuery.Execute(r.Context(), &getproduct.GetProductRequest{
		ProductID: id,
	})
	if err != nil {
		s.p.Log.Sugar().Errorw("getProduct", "id", id, "error", err)
		writeError(w, domainErrToStatus(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto)
}

// ── List ──────────────────────────────────────────────────────────────────────

func (s *Server) handleListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	req := &listproducts.ListProductsRequest{
		Limit:  parseIntParam(q.Get("limit"), 20),
		Offset: parseIntParam(q.Get("offset"), 0),
	}

	if cat := q.Get("category"); cat != "" {
		req.Category = &cat
	}

	resp, err := s.p.ListProductsQuery.Execute(r.Context(), req)
	if err != nil {
		s.p.Log.Sugar().Errorw("listProducts", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}
