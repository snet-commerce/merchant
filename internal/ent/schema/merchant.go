package schema

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	ment "github.com/snet-commerce/merchant/internal/ent"
	"github.com/snet-commerce/merchant/internal/ent/hook"
)

type Merchant struct {
	ent.Schema
}

func (Merchant) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(128).NotEmpty(),
		field.Text("description").Nillable().Optional(),
		field.String("primary_email").MaxLen(254).NotEmpty(),
		field.String("secondary_email").MaxLen(254).Nillable().Optional(),
		field.String("primary_number").MaxLen(16).NotEmpty(),
		field.String("secondary_number").MaxLen(16).Nillable().Optional(),
		field.UUID("tenant", uuid.UUID{}).Immutable().Unique(),
		field.Bool("active").Default(true),
	}
}

func (Merchant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		GUID{},
		ManagedAtMixin{},
	}
}

func (Merchant) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(func(mutator ent.Mutator) ent.Mutator {
			return hook.MerchantFunc(func(ctx context.Context, mut *ment.MerchantMutation) (ment.Value, error) {
				mut.Fields()
			})
		}, ent.OpCreate|ent.OpUpdate),
	}
}
