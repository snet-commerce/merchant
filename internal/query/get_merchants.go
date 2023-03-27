package query

import (
	"github.com/google/uuid"
	"github.com/snet-commerce/merchant/internal/ent"
	"github.com/snet-commerce/merchant/internal/ent/merchant"
	"github.com/snet-commerce/merchant/internal/ent/predicate"
)

type GetMerchantsParams struct {
	Name   *string
	Email  *string
	Number *string
	Active *bool
	Tenant *uuid.UUID
	Limit  int
	Offset int
}

type getMerchants struct {
	query *ent.MerchantQuery
}

func GetMerchants(q *ent.MerchantQuery) *getMerchants {
	return &getMerchants{query: q}
}

func (q *getMerchants) Apply(params GetMerchantsParams) *ent.MerchantQuery {
	predicates := make([]predicate.Merchant, 0)

	if params.Name != nil {
		predicates = append(predicates, merchant.NameEQ(*params.Name))
	}

	if params.Email != nil {
		predicates = append(predicates, merchant.Or(merchant.PrimaryEmailEQ(*params.Email), merchant.SecondaryEmailEQ(*params.Email)))
	}

	if params.Number != nil {
		predicates = append(predicates, merchant.Or(merchant.PrimaryNumberEQ(*params.Number), merchant.SecondaryNumberEQ(*params.Number)))
	}

	if params.Active != nil {
		predicates = append(predicates, merchant.ActiveEQ(*params.Active))
	}

	if params.Tenant != nil {
		predicates = append(predicates, merchant.TenantEQ(*params.Tenant))
	}

	return q.query.
		Where(predicates...).
		Limit(params.Limit).
		Offset(params.Offset).
		Order(ent.Asc())
}
