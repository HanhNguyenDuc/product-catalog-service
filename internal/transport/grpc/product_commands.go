package grpctransport

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	productv1 "github.com/product-catalog-service/gen/product/v1"
	activateproduct "github.com/product-catalog-service/internal/app/product/usecases/activate_product"
	applydiscount "github.com/product-catalog-service/internal/app/product/usecases/apply_discount"
	createproduct "github.com/product-catalog-service/internal/app/product/usecases/create_product"
	deactivateproduct "github.com/product-catalog-service/internal/app/product/usecases/deactivate_product"
	removediscount "github.com/product-catalog-service/internal/app/product/usecases/remove_discount"
	updateproduct "github.com/product-catalog-service/internal/app/product/usecases/update_product"
)

func (s *ProductServiceServer) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductReply, error) {
	id, err := s.p.CreateProductInteractor.Execute(ctx, &createproduct.CreateProductRequest{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
	})
	if err != nil {
		return nil, toStatusErr(err)
	}
	return &productv1.CreateProductReply{Id: id}, nil
}

func (s *ProductServiceServer) UpdateProduct(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductReply, error) {
	ucReq := &updateproduct.UpdateProductRequest{ProductID: req.Id}
	if req.Name != "" {
		ucReq.Name = &req.Name
	}
	if req.Description != "" {
		ucReq.Description = &req.Description
	}
	if req.Category != "" {
		ucReq.Category = &req.Category
	}

	if err := s.p.UpdateProductInteractor.Execute(ctx, ucReq); err != nil {
		return nil, toStatusErr(err)
	}
	return &productv1.UpdateProductReply{}, nil
}

func (s *ProductServiceServer) ActivateProduct(ctx context.Context, req *productv1.ActivateProductRequest) (*productv1.ActivateProductReply, error) {
	if err := s.p.ActivateProductInteractor.Execute(ctx, &activateproduct.ActivateProductRequest{ProductID: req.Id}); err != nil {
		return nil, toStatusErr(err)
	}
	return &productv1.ActivateProductReply{}, nil
}

func (s *ProductServiceServer) DeactivateProduct(ctx context.Context, req *productv1.DeactivateProductRequest) (*productv1.DeactivateProductReply, error) {
	if err := s.p.DeactivateProductInteractor.Execute(ctx, &deactivateproduct.DeactivateProductRequest{ProductID: req.Id}); err != nil {
		return nil, toStatusErr(err)
	}
	return &productv1.DeactivateProductReply{}, nil
}

func (s *ProductServiceServer) ApplyDiscount(ctx context.Context, req *productv1.ApplyDiscountRequest) (*productv1.ApplyDiscountReply, error) {
	if err := s.p.ApplyDiscountInteractor.Execute(ctx, &applydiscount.ApplyDiscountRequest{
		ProductID:  req.Id,
		Percentage: req.Percentage,
		StartsAt:   req.StartsAt.AsTime(),
		EndsAt:     req.EndsAt.AsTime(),
	}); err != nil {
		return nil, toStatusErr(err)
	}
	return &productv1.ApplyDiscountReply{}, nil
}

func (s *ProductServiceServer) RemoveDiscount(ctx context.Context, req *productv1.RemoveDiscountRequest) (*productv1.RemoveDiscountReply, error) {
	if err := s.p.RemoveDiscountInteractor.Execute(ctx, &removediscount.RemoveDiscountRequest{ProductID: req.Id}); err != nil {
		return nil, toStatusErr(err)
	}
	return &productv1.RemoveDiscountReply{}, nil
}

// helper — convert *timestamppb.Timestamp to proto (silences unused import).
var _ = timestamppb.Now
