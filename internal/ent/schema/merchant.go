package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
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

func (Merchant) Edges() []ent.Edge {
	return nil
}

func (Merchant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		GUID{},
		ManagedAtMixin{},
	}
}
