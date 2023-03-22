package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type ManagedAtMixin struct {
	mixin.Schema
}

func (ManagedAtMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Immutable().Default(time.Now().UTC).SchemaType(SchemaTypeTimestamp()),
		field.Time("updated_at").Default(time.Now().UTC).UpdateDefault(time.Now().UTC).SchemaType(SchemaTypeTimestamp()),
	}
}
