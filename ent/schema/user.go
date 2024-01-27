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
		field.String("username"),
		field.String("email"),
		field.String("pubKey").Unique(),
		field.String("image").Default("https://cdn.discordapp.com/attachments/876848373720842260/1200728031451414560/7748169_1.png?ex=65c73c1f&is=65b4c71f&hm=5653d203db5ac73730d8e2149dcd4bfc7c61bc90a1a9194cc8e07da3e86848de&"),
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
