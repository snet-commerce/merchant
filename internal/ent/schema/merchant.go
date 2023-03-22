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
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").MaxLen(128).NotEmpty(),
		field.Text("description").Nillable(),
		field.String("primary_email").MaxLen(254).NotEmpty(),
		field.String("secondary_email").MaxLen(254).Nillable(),
		field.String("primary_number").MaxLen(16).NotEmpty(),
		field.String("secondary_number").MaxLen(16).Nillable(),
		field.UUID("tenant", uuid.UUID{}).Immutable().Unique(),
		field.Bool("active").Default(true),
	}
}

func (Merchant) Edges() []ent.Edge {
	return nil
}

func (Merchant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ManagedAtMixin{},
	}
}
