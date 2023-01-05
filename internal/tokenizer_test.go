package internal

import (
	"bufio"
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanIdentifier(t *testing.T) {
	var tests = []struct {
		input string
		err   error
	}{
		{
			input: "name",
			err:   nil,
		},
		{
			input: "my_identifier",
			err:   nil,
		},
		{
			input: "my_identifier_with_123403",
			err:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokenizer := NewTokenizer(bufio.NewReader(bytes.NewBufferString(tt.input)))
			err := tokenizer.scan()
			require.Nil(t, err)

			ident, err := tokenizer.scanIdentifier()
			require.Nil(t, err)
			require.Equal(t, tt.input, ident)
		})
	}
}

func TestScanNumber(t *testing.T) {
	var tests = []struct {
		input  string
		expect string
		tok    Token
		err    error
	}{
		{
			input: "128",
			tok:   Token_Int,
			err:   nil,
		},
		{
			input: "128.65",
			tok:   Token_Float,
			err:   nil,
		},
		{
			input: "128.6",
			tok:   Token_Float,
			err:   nil,
		},
		{
			input:  "128.6.",
			expect: "128.6",
			tok:    Token_Float,
			err:    nil,
		},
		{
			input:  "128..6",
			expect: "128.",
			tok:    Token_Float,
			err:    nil,
		},
		{
			input:  "128a",
			expect: "128",
			tok:    Token_Int,
			err:    nil,
		},
		{
			input:  "128.a",
			expect: "128.",
			tok:    Token_Float,
			err:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokenizer := NewTokenizer(bufio.NewReader(bytes.NewBufferString(tt.input)))
			err := tokenizer.scan()
			require.Nil(t, err)
			tok, lit, err := tokenizer.scanNumber()
			require.Nil(t, err)
			require.Equal(t, tt.tok.String(), tok.String())

			expect := tt.expect
			if expect == "" {
				expect = tt.input
			}
			require.Equal(t, expect, lit)
		})
	}
}

func TestScanString(t *testing.T) {
	var tests = []struct {
		input  string
		expect string
		err    error
	}{
		{
			input:  `"hello world"`,
			expect: "hello world",
			err:    nil,
		},
		{
			input: `"string`,
			err:   errors.New("1:7 -> string literal expects to be closed with the \" character"),
		},
		{
			input:  `"it accepts any character @|¢∞¬÷ int keyword, struct. :"`,
			expect: "it accepts any character @|¢∞¬÷ int keyword, struct. :",
		},
		{
			input:  `"''"`,
			expect: "''",
		},
		{
			input:  `"print a quote: \""`,
			expect: `print a quote: "`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokenizer := NewTokenizer(bufio.NewReader(bytes.NewBufferString(tt.input)))
			err := tokenizer.scan() // scan "
			require.Nil(t, err)

			lit, err := tokenizer.scanString()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, lit)
		})
	}
}

func TestScanComment(t *testing.T) {
	var tests = []struct {
		input  string
		expect string
		err    error
	}{
		{
			input:  `//-styled comment`,
			expect: "-styled comment",
			err:    nil,
		},
		{
			input:  `/* -styled comment */`,
			expect: "-styled comment",
			err:    nil,
		},
		{
			input:  `/*** more than one * */`,
			expect: "** more than one *",
			err:    nil,
		},
		{
			input:  `/// more than one /`,
			expect: "/ more than one /",
			err:    nil,
		},
		{
			input: `/* oops i missed the end`,
			err:   errors.New("1:24 -> comment not terminated"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokenizer := NewTokenizer(bufio.NewReader(bytes.NewBufferString(tt.input)))
			err := tokenizer.scan() // scan /
			require.Nil(t, err)

			lit, err := tokenizer.scanComment()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, lit)
		})
	}
}
