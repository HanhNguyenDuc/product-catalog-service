package rest

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/product-catalog-service/internal/app/product/domain"
	getproduct "github.com/product-catalog-service/internal/app/product/queries/get_product"
	listproducts "github.com/product-catalog-service/internal/app/product/queries/list_products"
	activateproduct "github.com/product-catalog-service/internal/app/product/usecases/activate_product"
	applydiscount "github.com/product-catalog-service/internal/app/product/usecases/apply_discount"
	createproduct "github.com/product-catalog-service/internal/app/product/usecases/create_product"
	updateproduct "github.com/product-catalog-service/internal/app/product/usecases/update_product"
)

// Params bundles all handler dependencies injected by FX.
type Params struct {
	fx.In

	Log                       *zap.Logger
	CreateProductInteractor   *createproduct.CreateProductInteractor
	UpdateProductInteractor   *updateproduct.UpdateProductInteractor
	ApplyDiscountInteractor   *applydiscount.ApplyDiscountInteractor
	ActivateProductInteractor *activateproduct.ActivateProductInteractor
	GetProductQuery           *getproduct.GetProductQuery
	ListProductsQuery         *listproducts.ListProductsQuery
}

// Server holds the HTTP mux and handler dependencies.
type Server struct {
	Mux *http.ServeMux
	log *zap.Logger
	p   Params
}

// NewServer registers all routes and returns a Server ready to embed in http.Server.
func NewServer(p Params) *Server {
	s := &Server{Mux: http.NewServeMux(), log: p.Log, p: p}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// Health
	s.Mux.HandleFunc("GET /healthz", s.handleHealthz)

	// Write endpoints
	s.Mux.HandleFunc("POST /products", s.handleCreateProduct)
	s.Mux.HandleFunc("PUT /products/{id}", s.handleUpdateProduct)
	s.Mux.HandleFunc("POST /products/{id}/activate", s.handleActivateProduct)
	s.Mux.HandleFunc("POST /products/{id}/discount", s.handleApplyDiscount)

	// Read endpoints
	s.Mux.HandleFunc("GET /products/{id}", s.handleGetProduct)
	s.Mux.HandleFunc("GET /products", s.handleListProducts)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// NewHTTPServer creates an *http.Server with proper timeouts and FX lifecycle hooks.
func NewHTTPServer(lc fx.Lifecycle, srv *Server, log *zap.Logger, addr string) *http.Server {
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: srv.Mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("starting HTTP server", zap.String("addr", addr))
			go func() {
				if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error("HTTP server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("shutting down HTTP server")
			return httpSrv.Shutdown(ctx)
		},
	})

	return httpSrv
}

// domainErrToStatus maps domain sentinel errors to HTTP status codes.
func domainErrToStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrProductNotActive),
		errors.Is(err, domain.ErrProductNameRequired),
		errors.Is(err, domain.ErrProductBasePriceRequired),
		errors.Is(err, domain.ErrDiscountInvalidPercentage),
		errors.Is(err, domain.ErrDiscountInvalidPeriod),
		errors.Is(err, domain.ErrNoActiveDiscount),
		errors.Is(err, domain.ErrInvalidDiscountPeriod):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
