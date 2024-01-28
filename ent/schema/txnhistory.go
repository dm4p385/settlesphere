package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// TxnHistory holds the schema definition for the TxnHistory entity.
type TxnHistory struct {
	ent.Schema
}

// Fields of the TxnHistory.
func (TxnHistory) Fields() []ent.Field {
	return []ent.Field{
		field.Float("amount").Positive(),
		field.String("note").Optional(),
		field.Bool("settled").Default(false),
		field.Time("created_at").
			Default(time.Now),
		field.Time("settled_at").
			Nillable(),
	}
}

// Edges of the TxnHistory.
func (TxnHistory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("belongs_to", Group.Type).
			Ref("txn_history").
			Unique(),
		edge.From("source", User.Type).
			Ref("lent_history").
			Unique(),
		edge.To("destination", User.Type).
			Unique(),
	}
}
