package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// URLScalar repräsentiert einen URL-Scalar-Typ, der intern als String behandelt wird.
var URLScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "URL",
	Description: "Der URL-Scalar-Typ repräsentiert eine URL als String.",
	Serialize: func(value interface{}) interface{} {
		if s, ok := value.(string); ok {
			return s
		}
		return nil
	},
	ParseValue: func(value interface{}) interface{} {
		if s, ok := value.(string); ok {
			return s
		}
		return nil
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		if strVal, ok := valueAST.(*ast.StringValue); ok {
			return strVal.Value
		}
		return nil
	},
})
