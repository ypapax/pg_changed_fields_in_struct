package main

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_getStructName(t *testing.T) {
	r := require.New(t)
	r.Equal("IncomeStatement", getStructName("type IncomeStatement struct {"))
}

func Test_parseField(t *testing.T) {
	r := require.New(t)
	b, err := os.ReadFile("parseField.test")
	r.NoError(err)
	r.NotEmpty(b)
	a, err := parseField(string(b))
	r.NoError(err)
	r.Equal(&Field{Name: "ID", Pg: "id"}, a)
}

func Test_parseStruct(t *testing.T) {
	r := require.New(t)
	b, err := os.ReadFile("parseStruct.test")
	r.NoError(err)
	r.NotEmpty(b)
	a, err := parseStruct(string(b))
	t.Logf("actual: %+v", a)
	r.NoError(err)
	bExp, err := os.ReadFile("parseStruct.expected.test")
	r.NoError(err)
	r.NotEmpty(bExp)
	r.Equal(string(bExp), a)
}
