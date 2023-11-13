package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Unique(),
		field.String("email").Unique(),
		field.String("pubKey").Unique(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("member_of", Group.Type).
			Ref("users"),
		edge.To("lent", Transaction.Type),
		edge.To("lent_history", TxnHistory.Type),
		edge.From("owed", Transaction.Type).
			Ref("destination"),
		edge.From("owed_history", TxnHistory.Type).
			Ref("destination"),
	}
}
