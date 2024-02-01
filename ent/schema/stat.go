package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Stat holds the schema definition for the Stat entity.
type Stat struct {
	ent.Schema
}

// Fields of the Stat.
func (Stat) Fields() []ent.Field {
	return []ent.Field{
		field.Float("total_paid").Positive().Default(0),
		field.Float("total_share").Positive().Default(0),
	}
}

// Edges of the Stat.
func (Stat) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("belongs_to_group", Group.Type).Unique(),
		edge.To("belongs_to_user", User.Type).Unique(),
	}
}
