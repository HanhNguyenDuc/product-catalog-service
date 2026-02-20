// Package appservices wires all application dependencies as a single fx.Option
// that can be passed directly to fx.New in main.
package appservices

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"cloud.google.com/go/spanner"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/product-catalog-service/common"
	"github.com/product-catalog-service/common/commitplanner"
	"github.com/product-catalog-service/internal/app/product/contract"
	"github.com/product-catalog-service/internal/app/product/domain/services"
	getproduct "github.com/product-catalog-service/internal/app/product/queries/get_product"
	listproducts "github.com/product-catalog-service/internal/app/product/queries/list_products"
	"github.com/product-catalog-service/internal/app/product/repo"
	activateproduct "github.com/product-catalog-service/internal/app/product/usecases/activate_product"
	applydiscount "github.com/product-catalog-service/internal/app/product/usecases/apply_discount"
	createproduct "github.com/product-catalog-service/internal/app/product/usecases/create_product"
	deactivateproduct "github.com/product-catalog-service/internal/app/product/usecases/deactivate_product"
	removediscount "github.com/product-catalog-service/internal/app/product/usecases/remove_discount"
	updateproduct "github.com/product-catalog-service/internal/app/product/usecases/update_product"
	grpctransport "github.com/product-catalog-service/internal/transport/grpc"
	"github.com/product-catalog-service/internal/transport/rest"
)

// CommonOptions provides all domain logic, repositories, and valid use cases.
// It does NOT include any transport layer.
var CommonOptions = fx.Options(
	// ── Infrastructure ───────────────────────────────────────────────────────
	fx.Provide(
		newLogger,
		newSpannerClient,
		newCommitter,
		newTicker,
	),

	// ── Repositories ─────────────────────────────────────────────────────────
	fx.Provide(
		fx.Annotate(
			newProductRepo,
			fx.As(new(contract.ProductRepository)),
			fx.As(new(contract.QueryRepository)),
		),
		fx.Annotate(
			newEventRepo,
			fx.As(new(contract.EventRepository)),
		),
	),

	// ── Domain services ───────────────────────────────────────────────────────
	fx.Provide(
		services.NewPricingCalculator,
	),

	// ── Use cases ─────────────────────────────────────────────────────────────
	fx.Provide(
		createproduct.NewCreateProductInteractor,
		updateproduct.NewUpdateProductInteractor,
		applydiscount.NewApplyDiscountInteractor,
		activateproduct.NewActivateProductInteractor,
		deactivateproduct.NewDeactivateProductInteractor,
		removediscount.NewRemoveDiscountInteractor,
	),

	// ── Queries ───────────────────────────────────────────────────────────────
	fx.Provide(
		getproduct.NewGetProductQuery,
		listproducts.NewListProductsQuery,
	),
)

// HTTPOptions plugs the REST transport layer on top of CommonOptions.
var HTTPOptions = fx.Options(
	fx.Provide(
		fx.Annotate(newHTTPAddr, fx.ResultTags(`name:"http_addr"`)),
		rest.NewServer,
		fx.Annotate(rest.NewHTTPServer, fx.ParamTags(``, ``, ``, `name:"http_addr"`)),
	),
	fx.Invoke(func(*http.Server) {}),
)

// GRPCOptions plugs the gRPC transport layer on top of CommonOptions.
var GRPCOptions = fx.Options(
	fx.Provide(
		fx.Annotate(newGRPCAddr, fx.ResultTags(`name:"grpc_addr"`)),
		grpctransport.NewProductServiceServer,
		fx.Annotate(grpctransport.NewGRPCServer, fx.ParamTags(``, ``, ``, `name:"grpc_addr"`)),
	),
	fx.Invoke(func(*grpc.Server) {}),
)

// ── Infrastructure constructors ───────────────────────────────────────────────

func newLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

func newSpannerClient(lc fx.Lifecycle, log *zap.Logger) (*spanner.Client, error) {
	// Map SPANNER_ENDPOINT to Google's standard SPANNER_EMULATOR_HOST
	endpoint := os.Getenv("SPANNER_ENDPOINT")
	if endpoint != "" {
		os.Setenv("SPANNER_EMULATOR_HOST", endpoint)
	}

	dsn := os.Getenv("SPANNER_DSN")
	if dsn == "" {
		project := os.Getenv("SPANNER_PROJECT")
		if project == "" {
			project = "local"
		}
		instance := os.Getenv("SPANNER_INSTANCE")
		if instance == "" {
			instance = "dev"
		}
		database := os.Getenv("SPANNER_DATABASE")
		if database == "" {
			database = "product-catalog"
		}
		dsn = fmt.Sprintf("projects/%s/instances/%s/databases/%s", project, instance, database)
	}

	client, err := spanner.NewClient(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("closing Spanner client")
			client.Close()
			return nil
		},
	})

	return client, nil
}

func newCommitter(client *spanner.Client) commitplanner.Applier {
	return commitplanner.NewCommitter(client)
}

func newTicker() common.Ticker {
	return common.NewRealTicker()
}

func newHTTPAddr() string {
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		return addr
	}
	return ":8080"
}

func newGRPCAddr() string {
	if addr := os.Getenv("GRPC_ADDR"); addr != "" {
		return addr
	}
	return ":50051"
}

func newProductRepo(client *spanner.Client) *repo.ProductRepo {
	return repo.NewProductRepo(client)
}

func newEventRepo() *repo.EventRepo {
	return repo.NewEventRepo()
}
