package main

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "single string",
			input:    []string{"hello"},
			expected: "hello",
		},
		{
			name:     "multiple strings",
			input:    []string{"hello", " ", "world"},
			expected: "hello world",
		},
		{
			name:     "empty strings",
			input:    []string{"", "", ""},
			expected: "",
		},
		{
			name:     "mixed strings",
			input:    []string{"foo", "", "bar"},
			expected: "foobar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b builder
			b.WriteStrings(tt.input...)
			assert.Equal(t, tt.expected, b.String())
		})
	}
}
func TestNewRand(t *testing.T) {
	tests := []struct {
		name  string
		seeds []uint64
	}{
		{
			name:  "single seed",
			seeds: []uint64{12345},
		},
		{
			name:  "multiple seeds",
			seeds: []uint64{12345, 67890, 54321},
		},
		{
			name:  "no seeds",
			seeds: []uint64{},
		},
		{
			name:  "all zero seeds",
			seeds: []uint64{0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRand(tt.seeds)
			assert.NotNil(t, r)
			assert.IsType(t, &rand.Rand{}, r)
		})
	}
}
func TestNewRandDeterminism(t *testing.T) {
	r1 := newRand([]uint64{12345, 67890})
	r2 := newRand([]uint64{12345, 67890})
	assert.Equal(t, r1.Uint64(), r2.Uint64(), "newRand should be deterministic for same seeds")
}

func TestPiaFormat(t *testing.T) {
	seed := []uint64{1}
	r := newRand(seed)
	query := "hello"
	ctx := &UpdateContext{
		Rand:  r,
		Query: &query,
	}
	result := pia(ctx)
	assert.True(t, strings.HasPrefix(result, "Pia!"), "pia should start with Pia! prefix")
	assert.True(t, strings.HasSuffix(result, "hello"), "pia should end with query")
}

func TestDivineOutput(t *testing.T) {
	seed := []uint64{2}
	r := newRand(seed)
	query := "question"
	ctx := &UpdateContext{
		Rand:  r,
		Query: &query,
	}
	result := divine(ctx)
	assert.Contains(t, result, "所求事项: question", "divine should contain query")
	assert.Contains(t, result, "结果: ", "divine should contain 结果: ")
}
