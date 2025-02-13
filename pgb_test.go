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
func TestPia(t *testing.T) {
	const n = 10000
	const seed = 42

	expTotal := 8
	expRatios := map[string]int{
		"Pia!▼(ｏ ‵-′)ノ★":  1,
		"Pia!<(=ｏ ‵-′)ノ☆": 7,
	}

	counts := make(map[string]int)
	for k := range expRatios {
		counts[k] = 0
	}

	r := rand.New(rand.NewSource(seed))
	for i := 0; i < n; i++ {
		query := "test query"
		ctx := &UpdateContext{
			Rand:  r,
			Query: &query,
		}
		result := pia(ctx)
		for key := range counts {
			if strings.Contains(result, key) {
				counts[key]++
				break
			}
		}
	}

	for k, v := range counts {
		exp := float64(expRatios[k]) / float64(expTotal) * n
		tol := exp * 2
		t.Logf("%s: %d, expected: %f, actual: %d, tolarence: %f", k, v, exp, v, tol)
		assert.InDelta(t, exp, v, tol, "Ratio mismatch for: "+k)
	}
}

func TestDivine(t *testing.T) {
	const n = 10000
	const seed = 42

	expTotal := 16 * 1024
	expRatios := map[string]int{
		"极大吉": 7 * 1,
		"超大吉": 7 * 10,
		"特大吉": 7 * 45,
		"甚大吉": 7 * 120,
		"大吉":  7 * 210,
		"吉":   7 * 252,
		"小吉":  7 * 210,
		"甚小吉": 7 * 120,
		"特小吉": 7 * 45,
		"超小吉": 7 * 10,
		"极小吉": 7 * 1,
		"尚可":  2 * 1024,
		"极小凶": 7 * 1,
		"超小凶": 7 * 10,
		"特小凶": 7 * 45,
		"甚小凶": 7 * 120,
		"小凶":  7 * 210,
		"凶":   7 * 252,
		"大凶":  7 * 210,
		"甚大凶": 7 * 120,
		"特大凶": 7 * 45,
		"超大凶": 7 * 10,
		"极大凶": 7 * 1,
	}

	counts := make(map[string]int)
	for k := range expRatios {
		counts[k] = 0
	}

	r := rand.New(rand.NewSource(seed))
	for i := 0; i < n; i++ {
		query := "test query"
		ctx := &UpdateContext{
			Rand:  r,
			Query: &query,
		}
		result := divine(ctx)
		for key := range counts {
			if strings.Contains(result, key) {
				counts[key]++
				break
			}
		}
	}

	for k, v := range counts {
		exp := float64(expRatios[k]) / float64(expTotal) * n
		tol := exp * 3
		t.Logf("%s: %d, expected: %f, actual: %d, tolarence: %f", k, v, exp, v, tol)
		assert.InDelta(t, exp, v, tol, "Ratio mismatch for: "+k)
	}
}
