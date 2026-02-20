package grpctransport

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	productv1 "github.com/product-catalog-service/gen/product/v1"
	getproduct "github.com/product-catalog-service/internal/app/product/queries/get_product"
	listproducts "github.com/product-catalog-service/internal/app/product/queries/list_products"
)

func (s *ProductServiceServer) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductReply, error) {
	dto, err := s.p.GetProductQuery.Execute(ctx, &getproduct.GetProductRequest{ProductID: req.Id})
	if err != nil {
		return nil, toStatusErr(err)
	}
	return &productv1.GetProductReply{Product: toProtoProduct(dto)}, nil
}

func (s *ProductServiceServer) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsReply, error) {
	ucReq := &listproducts.ListProductsRequest{
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}
	if req.Category != "" {
		ucReq.Category = &req.Category
	}

	resp, err := s.p.ListProductsQuery.Execute(ctx, ucReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	products := make([]*productv1.Product, 0, len(resp.Items))
	for _, item := range resp.Items {
		products = append(products, toProtoProductSummary(item))
	}

	return &productv1.ListProductsReply{
		Products:   products,
		TotalCount: int32(resp.TotalCount),
	}, nil
}

// ── Mapping helpers ───────────────────────────────────────────────────────────

func toProtoProduct(dto *getproduct.ProductDTO) *productv1.Product {
	p := &productv1.Product{
		Id:             dto.ID,
		Name:           dto.Name,
		Description:    dto.Description,
		Category:       dto.Category,
		Status:         dto.Status,
		BasePrice:      toProtoMoney(dto.BasePrice.Amount, dto.BasePrice.Currency),
		EffectivePrice: toProtoMoney(dto.EffectivePrice.Amount, dto.EffectivePrice.Currency),
	}
	if dto.Discount != nil {
		p.Discount = &productv1.Discount{
			AmountPercentage: dto.Discount.Percentage,
			StartsAt:         timestamppb.New(dto.Discount.StartsAt),
			EndsAt:           timestamppb.New(dto.Discount.EndsAt),
			IsActive:         dto.Discount.IsActive,
		}
	}
	return p
}

func toProtoProductSummary(dto *listproducts.ProductSummaryDTO) *productv1.Product {
	p := &productv1.Product{
		Id:             dto.ID,
		Name:           dto.Name,
		Category:       dto.Category,
		Status:         dto.Status,
		BasePrice:      toProtoMoney(dto.BasePrice.Amount, dto.BasePrice.Currency),
		EffectivePrice: toProtoMoney(dto.EffectivePrice.Amount, dto.EffectivePrice.Currency),
	}
	if dto.DiscountEndsAt != nil {
		p.Discount = &productv1.Discount{
			IsActive: dto.IsDiscounted,
			EndsAt:   timestamppb.New(*dto.DiscountEndsAt),
		}
	}
	return p
}

func toProtoMoney(amount int64, currency string) *productv1.Money {
	return &productv1.Money{Amount: amount, Currency: currency}
}
