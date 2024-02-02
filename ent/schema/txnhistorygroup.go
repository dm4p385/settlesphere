package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
)

// TxnHistoryGroup holds the schema definition for the TxnHistoryGroup entity.
type TxnHistoryGroup struct {
	ent.Schema
}

// Fields of the TxnHistoryGroup.
func (TxnHistoryGroup) Fields() []ent.Field {
	return nil
}

// Edges of the TxnHistoryGroup.
func (TxnHistoryGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("txn_history", TxnHistory.Type),
	}
}
