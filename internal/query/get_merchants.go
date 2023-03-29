package query

import (
	"github.com/google/uuid"
	"github.com/snet-commerce/merchant/internal/ent"
	"github.com/snet-commerce/merchant/internal/ent/merchant"
	"github.com/snet-commerce/merchant/internal/ent/predicate"
)

type GetMerchantsQueryParams struct {
	Name   *string
	Email  *string
	Number *string
	Active *bool
	Tenant *uuid.UUID
	Limit  int
	Offset int
}

type GetMerchantsQuery struct {
	query *ent.MerchantQuery
}

func GetMerchants(q *ent.MerchantQuery) *GetMerchantsQuery {
	return &GetMerchantsQuery{query: q}
}

func (q *GetMerchantsQuery) Apply(params GetMerchantsQueryParams) *ent.MerchantQuery {
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
