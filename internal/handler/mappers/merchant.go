package mappers

import (
	pb "buf.build/gen/go/snet-commerce/merchant/protocolbuffers/go/merchant/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/snet-commerce/merchant/internal/ent"
)

func MerchantToProtobuf(m *ent.Merchant) *pb.Merchant {
	return &pb.Merchant{
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
	}
}

func MerchantsToProtobuf(merchants []*ent.Merchant) []*pb.Merchant {
	data := make([]*pb.Merchant, 0, len(merchants))
	for i := range merchants {
		data = append(data, MerchantToProtobuf(merchants[i]))
	}
	return data
}
