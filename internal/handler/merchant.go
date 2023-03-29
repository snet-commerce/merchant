package handler

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/snet-commerce/merchant/internal/ent"
	"github.com/snet-commerce/merchant/internal/handler/mappers"
	"github.com/snet-commerce/merchant/internal/query"
	pb "github.com/snet-commerce/merchant/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MerchantHandler struct {
	client *ent.MerchantClient
	logger *zap.SugaredLogger
	pb.UnimplementedMerchantServiceServer
}

func NewMerchantHandler(
	client *ent.MerchantClient,
	logger *zap.SugaredLogger,
) *MerchantHandler {
	return &MerchantHandler{client: client, logger: logger}
}

func (h *MerchantHandler) CreateMerchant(ctx context.Context, req *pb.CreateMerchantRequest) (*pb.CreateMerchantResponse, error) {
	m, err := h.client.
		Create().
		SetID(uuid.New()).
		SetName(req.Name).
		SetNillableDescription(req.Description).
		SetPrimaryEmail(req.PrimaryEmail).
		SetNillableSecondaryEmail(req.SecondaryEmail).
		SetPrimaryNumber(req.PrimaryNumber).
		SetNillableSecondaryNumber(req.SecondaryNumber).
		SetTenant(uuid.New()).
		SetActive(req.Active).
		Save(ctx)
	if err != nil {
		h.logger.Errorf("failed to create merchant - %s", err)
		return nil, err
	}

	return &pb.CreateMerchantResponse{Merchant: mappers.MerchantToProtobuf(m)}, nil
}

func (h *MerchantHandler) GetMerchant(ctx context.Context, req *pb.GetMerchantRequest) (*pb.GetMerchantResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to parse id as uuid - %s", err))
	}

	m, err := h.client.Get(ctx, id)
	if err != nil {
		h.logger.Errorf("failed to read merchant - %s", err)
		return nil, err
	}

	return &pb.GetMerchantResponse{Merchant: mappers.MerchantToProtobuf(m)}, nil
}

func (h *MerchantHandler) UpdateMerchant(ctx context.Context, req *pb.UpdateMerchantRequest) (*pb.UpdateMerchantResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to parse id as uuid - %s", err))
	}

	m, err := h.client.
		UpdateOneID(id).
		SetName(req.Name).
		SetNillableDescription(req.Description).
		SetPrimaryEmail(req.PrimaryEmail).
		SetNillableSecondaryEmail(req.SecondaryEmail).
		SetPrimaryNumber(req.PrimaryNumber).
		SetNillableSecondaryNumber(req.SecondaryNumber).
		SetActive(req.Active).
		Save(ctx)
	if err != nil {
		h.logger.Errorf("failed to update merchant - %s", err)
		return nil, err
	}

	return &pb.UpdateMerchantResponse{Merchant: mappers.MerchantToProtobuf(m)}, nil
}

func (h *MerchantHandler) DeleteMerchant(ctx context.Context, req *pb.DeleteMerchantRequest) (*pb.DeleteMerchantResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to parse id as uuid - %s", err))
	}

	if err := h.client.DeleteOneID(id).Exec(ctx); err != nil {
		h.logger.Errorf("failed to delete merchant - %s", err)
		return nil, err
	}

	return &pb.DeleteMerchantResponse{}, nil
}

func (h *MerchantHandler) GetMerchants(ctx context.Context, req *pb.GetMerchantsRequest) (*pb.GetMerchantsResponse, error) {
	var tenant *uuid.UUID
	if req.Tenant != nil {
		t, err := uuid.Parse(*req.Tenant)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to parse tenant id as uuid - %s", err))
		}
		tenant = &t
	}

	q := query.GetMerchants(h.client.Query()).Apply(
		query.GetMerchantsQueryParams{
			Name:   req.Name,
			Email:  req.Email,
			Number: req.Number,
			Active: req.Active,
			Tenant: tenant,
			Limit:  int(req.Limit),
			Offset: int(req.Offset),
		},
	)

	merchants, err := q.All(ctx)
	if err != nil {
		h.logger.Errorf("failed to read merchants - %s", err)
		return nil, err
	}

	return &pb.GetMerchantsResponse{Merchants: mappers.MerchantsToProtobuf(merchants)}, nil
}
