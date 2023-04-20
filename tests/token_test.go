package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
	"tomasweigenast.com/nexema/tool/internal/token"
)

func TestToken_ToKeyword(t *testing.T) {
	tests := []struct {
		input *token.Token
		want  *token.Token
	}{
		{token.NewToken(token.Ident, "as"), token.NewToken(token.As, "as")},
		{token.NewToken(token.Ident, "type"), token.NewToken(token.Type, "type")},
		{token.NewToken(token.Ident, "struct"), token.NewToken(token.Struct, "struct")},
		{token.NewToken(token.Ident, "enum"), token.NewToken(token.Enum, "enum")},
		{token.NewToken(token.Ident, "union"), token.NewToken(token.Union, "union")},
		{token.NewToken(token.Ident, "base"), token.NewToken(token.Base, "base")},
		{token.NewToken(token.Ident, "extends"), token.NewToken(token.Extends, "extends")},
		{token.NewToken(token.Ident, "defaults"), token.NewToken(token.Defaults, "defaults")},
		{token.NewToken(token.Ident, "use"), token.NewToken(token.Use, "use")},
		{token.NewToken(token.Ident, "let"), nil},
	}
	for _, tt := range tests {
		t.Run(tt.input.Literal, func(t *testing.T) {
			result := tt.input.ToKeyword()
			require.Equal(t, tt.want, result)
		})
	}
}
