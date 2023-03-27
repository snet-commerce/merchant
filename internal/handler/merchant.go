package handler

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/snet-commerce/merchant/internal/ent"
	"github.com/snet-commerce/merchant/internal/ent/merchant"
	"github.com/snet-commerce/merchant/internal/ent/predicate"
	pb "github.com/snet-commerce/merchant/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MerchantHandler struct {
	client *ent.Client
	logger *zap.SugaredLogger
	pb.UnimplementedMerchantServiceServer
}

func NewMerchantHandler(
	client *ent.Client,
	logger *zap.SugaredLogger,
) *MerchantHandler {
	return &MerchantHandler{client: client, logger: logger}
}

func (h *MerchantHandler) CreateMerchant(ctx context.Context, req *pb.CreateMerchantRequest) (*pb.CreateMerchantResponse, error) {
	m, err := h.client.Merchant.
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

	return &pb.CreateMerchantResponse{
		Merchant: &pb.Merchant{
			Id:              m.ID.String(),
			Name:            m.Name,
			Description:     m.Description,
			PrimaryEmail:    m.PrimaryEmail,
			SecondaryEmail:  m.SecondaryEmail,
			PrimaryNumber:   m.PrimaryNumber,
			SecondaryNumber: m.SecondaryNumber,
			Active:          m.Active,
			Tenant:          m.Tenant.String(),
			CreatedAt:       timestamppb.New(m.CreatedAt),
			UpdatedAt:       timestamppb.New(m.UpdatedAt),
		},
	}, nil
}

func (h *MerchantHandler) GetMerchant(ctx context.Context, req *pb.GetMerchantRequest) (*pb.GetMerchantResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to parse id as uuid - %s", err))
	}

	m, err := h.client.Merchant.Get(ctx, id)
	if err != nil {
		h.logger.Errorf("failed to read merchant - %s", err)
		return nil, err
	}

	return &pb.GetMerchantResponse{
		Merchant: &pb.Merchant{
			Id:              m.ID.String(),
			Name:            m.Name,
			Description:     m.Description,
			PrimaryEmail:    m.PrimaryEmail,
			SecondaryEmail:  m.SecondaryEmail,
			PrimaryNumber:   m.PrimaryNumber,
			SecondaryNumber: m.SecondaryNumber,
			Active:          m.Active,
			Tenant:          m.Tenant.String(),
			CreatedAt:       timestamppb.New(m.CreatedAt),
			UpdatedAt:       timestamppb.New(m.UpdatedAt),
		},
	}, nil
}

func (h *MerchantHandler) UpdateMerchant(ctx context.Context, req *pb.UpdateMerchantRequest) (*pb.UpdateMerchantResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to parse id as uuid - %s", err))
	}

	m, err := h.client.Merchant.
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

	return &pb.UpdateMerchantResponse{
		Merchant: &pb.Merchant{
			Id:              m.ID.String(),
			Name:            m.Name,
			Description:     m.Description,
			PrimaryEmail:    m.PrimaryEmail,
			SecondaryEmail:  m.SecondaryEmail,
			PrimaryNumber:   m.PrimaryNumber,
			SecondaryNumber: m.SecondaryNumber,
			Active:          m.Active,
			Tenant:          m.Tenant.String(),
			CreatedAt:       timestamppb.New(m.CreatedAt),
			UpdatedAt:       timestamppb.New(m.UpdatedAt),
		},
	}, nil
}

func (h *MerchantHandler) DeleteMerchant(ctx context.Context, req *pb.DeleteMerchantRequest) (*pb.DeleteMerchantResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to parse id as uuid - %s", err))
	}

	if err := h.client.Merchant.DeleteOneID(id).Exec(ctx); err != nil {
		h.logger.Errorf("failed to delete merchant - %s", err)
		return nil, err
	}

	return &pb.DeleteMerchantResponse{}, nil
}

func (h *MerchantHandler) GetMerchants(ctx context.Context, req *pb.GetMerchantsRequest) (*pb.GetMerchantsResponse, error) {
	predicates := make([]predicate.Merchant, 0)

	if req.Name != nil {
		predicates = append(predicates, merchant.NameEQ(*req.Name))
	}

	if req.Email != nil {
		predicates = append(predicates, merchant.Or(merchant.PrimaryEmailEQ(*req.Email), merchant.SecondaryEmailEQ(*req.Email)))
	}

	if req.Number != nil {
		predicates = append(predicates, merchant.Or(merchant.PrimaryNumberEQ(*req.Number), merchant.SecondaryNumberEQ(*req.Number)))
	}

	if req.Active != nil {
		predicates = append(predicates, merchant.ActiveEQ(*req.Active))
	}

	if req.Tenant != nil {
		tenant, err := uuid.Parse(*req.Tenant)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to parse tenant id as uuid - %s", err))
		}
		predicates = append(predicates, merchant.TenantEQ(tenant))
	}

	merchants, err := h.client.Merchant.
		Query().
		Where(
			predicates...,
		).
		Limit(int(req.Limit)).
		Offset(int(req.Offset)).
		All(ctx)
	if err != nil {
		h.logger.Errorf("failed to read merchants - %s", err)
		return nil, err
	}

	res := make([]*pb.Merchant, 0, len(merchants))
	for i := range merchants {
		m := merchants[i]
		res = append(res, &pb.Merchant{
			Id:              m.ID.String(),
			Name:            m.Name,
			Description:     m.Description,
			PrimaryEmail:    m.PrimaryEmail,
			SecondaryEmail:  m.SecondaryEmail,
			PrimaryNumber:   m.PrimaryNumber,
			SecondaryNumber: m.SecondaryNumber,
			Active:          m.Active,
			Tenant:          m.Tenant.String(),
			CreatedAt:       timestamppb.New(m.CreatedAt),
			UpdatedAt:       timestamppb.New(m.UpdatedAt),
		})
	}

	return &pb.GetMerchantsResponse{Merchants: res}, nil
}
