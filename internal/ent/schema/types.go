package schema

import (
	"entgo.io/ent/dialect"
)

func SchemaTypeTimestamp() map[string]string {
	return map[string]string{dialect.Postgres: "timestamp"}
}
