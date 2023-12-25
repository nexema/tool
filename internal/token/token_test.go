package token

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToken_ToKeyword(t *testing.T) {
	tests := []struct {
		input *Token
		want  *Token
	}{
		{NewToken(Ident, "as"), NewToken(As, "as")},
		{NewToken(Ident, "type"), NewToken(Type, "type")},
		{NewToken(Ident, "struct"), NewToken(Struct, "struct")},
		{NewToken(Ident, "enum"), NewToken(Enum, "enum")},
		{NewToken(Ident, "union"), NewToken(Union, "union")},
		{NewToken(Ident, "base"), NewToken(Base, "base")},
		{NewToken(Ident, "extends"), NewToken(Extends, "extends")},
		{NewToken(Ident, "defaults"), NewToken(Defaults, "defaults")},
		{NewToken(Ident, "include"), NewToken(Include, "include")},
		{NewToken(Ident, "let"), nil},
	}
	for _, tt := range tests {
		t.Run(tt.input.Literal, func(t *testing.T) {
			result := tt.input.ToKeyword()
			require.Equal(t, tt.want, result)
		})
	}
}
