package grpctransport

import (
	"context"
	"errors"
	"net"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	productv1 "github.com/product-catalog-service/gen/product/v1"
	"github.com/product-catalog-service/internal/app/product/domain"
	getproduct "github.com/product-catalog-service/internal/app/product/queries/get_product"
	listproducts "github.com/product-catalog-service/internal/app/product/queries/list_products"
	activateproduct "github.com/product-catalog-service/internal/app/product/usecases/activate_product"
	applydiscount "github.com/product-catalog-service/internal/app/product/usecases/apply_discount"
	createproduct "github.com/product-catalog-service/internal/app/product/usecases/create_product"
	deactivateproduct "github.com/product-catalog-service/internal/app/product/usecases/deactivate_product"
	removediscount "github.com/product-catalog-service/internal/app/product/usecases/remove_discount"
	updateproduct "github.com/product-catalog-service/internal/app/product/usecases/update_product"
)

// Params bundles all handler dependencies injected by FX.
type Params struct {
	fx.In

	Log                         *zap.Logger
	CreateProductInteractor     *createproduct.CreateProductInteractor
	UpdateProductInteractor     *updateproduct.UpdateProductInteractor
	ActivateProductInteractor   *activateproduct.ActivateProductInteractor
	DeactivateProductInteractor *deactivateproduct.DeactivateProductInteractor
	ApplyDiscountInteractor     *applydiscount.ApplyDiscountInteractor
	RemoveDiscountInteractor    *removediscount.RemoveDiscountInteractor
	GetProductQuery             *getproduct.GetProductQuery
	ListProductsQuery           *listproducts.ListProductsQuery
}

// ProductServiceServer implements productv1.ProductServiceServer.
type ProductServiceServer struct {
	productv1.UnimplementedProductServiceServer
	p Params
}

// NewProductServiceServer creates and registers the gRPC server implementation.
func NewProductServiceServer(p Params) *ProductServiceServer {
	return &ProductServiceServer{p: p}
}

// NewGRPCServer starts a gRPC server with FX lifecycle management.
func NewGRPCServer(lc fx.Lifecycle, svc *ProductServiceServer, log *zap.Logger, addr string) *grpc.Server {
	srv := grpc.NewServer()
	productv1.RegisterProductServiceServer(srv, svc)
	reflection.Register(srv)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", addr)
			if err != nil {
				return err
			}
			log.Info("starting gRPC server", zap.String("addr", addr))
			go func() {
				if err := srv.Serve(lis); err != nil {
					log.Error("gRPC server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("shutting down gRPC server")
			srv.GracefulStop()
			return nil
		},
	})

	return srv
}

// domainErrToCode maps domain sentinel errors to gRPC status codes.
func domainErrToCode(err error) codes.Code {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return codes.NotFound
	case errors.Is(err, domain.ErrProductNameRequired),
		errors.Is(err, domain.ErrProductBasePriceRequired),
		errors.Is(err, domain.ErrDiscountInvalidPercentage),
		errors.Is(err, domain.ErrDiscountInvalidPeriod),
		errors.Is(err, domain.ErrInvalidDiscountPeriod),
		errors.Is(err, domain.ErrNoActiveDiscount):
		return codes.InvalidArgument
	case errors.Is(err, domain.ErrProductNotActive):
		return codes.FailedPrecondition
	default:
		return codes.Internal
	}
}

func toStatusErr(err error) error {
	return status.Error(domainErrToCode(err), err.Error())
}
