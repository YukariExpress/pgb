package main

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/go-telegram/bot/models"
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

func TestBuildInlineQueryResults(t *testing.T) {
	user := &models.User{
		ID:           42,
		IsBot:        false,
		FirstName:    "Test",
		LastName:     "",
		Username:     "",
		LanguageCode: "zh",
	}
	results := buildInlineQueryResults(user, "问题")
	assert.Equal(t, 2, len(results), "Should return 2 results")
	if article, ok := results[0].(*models.InlineQueryResultArticle); ok {
		assert.True(t, article.Title == "求签" || article.Title == "Divination", "First result title should be '求签' or 'Divination'")
	} else {
		t.Fatalf("First result is not an InlineQueryResultArticle")
	}
}

func TestGetLocaleTitles(t *testing.T) {
	d, p := getLocaleTitles("zh")
	assert.Equal(t, "求签", d)
	assert.Equal(t, "Pia", p)
	d, p = getLocaleTitles("en")
	assert.Equal(t, "Divination", d)
	assert.Equal(t, "Pia", p)
	d, p = getLocaleTitles("")
	assert.Equal(t, "Divination", d)
	assert.Equal(t, "Pia", p)
}

func TestBuildUpdateContext(t *testing.T) {
	userID := uint64(12345)
	query := "test-query"
	locale := "zh"
	rctx1 := buildUpdateContext(userID, query, locale)
	rctx2 := buildUpdateContext(userID, query, locale)
	assert.NotNil(t, rctx1)
	assert.NotNil(t, rctx2)
	assert.Equal(t, *rctx1.Query, query)
	assert.Equal(t, *rctx1.Locale, locale)
	// Deterministic: first random value should be the same
	v1 := rctx1.Rand.Uint64()
	v2 := rctx2.Rand.Uint64()
	assert.Equal(t, v1, v2, "Random context should be deterministic for same input")
}

func TestGetUserID(t *testing.T) {
	user := &models.User{ID: 12345}
	assert.Equal(t, uint64(12345), getUserID(user))
	assert.Equal(t, uint64(0), getUserID(nil))
}

func TestGetUserLocale(t *testing.T) {
	user := &models.User{LanguageCode: "en"}
	assert.Equal(t, "en", getUserLocale(user))
	user.LanguageCode = ""
	assert.Equal(t, "zh", getUserLocale(user))
	assert.Equal(t, "zh", getUserLocale(nil))
}

func TestGetPiaPrefix(t *testing.T) {
	tests := []struct {
		name      string
		randValue uint64
		expected  string
	}{
		{
			name:      "dog prefix (case 0)",
			randValue: 0,
			expected:  "Pia!▼(ｏ ‵-′)ノ★ ",
		},
		{
			name:      "dog prefix (case 8)",
			randValue: 8,
			expected:  "Pia!▼(ｏ ‵-′)ノ★ ",
		},
		{
			name:      "cat prefix (case 1)",
			randValue: 1,
			expected:  "Pia!<(=ｏ ‵-′)ノ☆ ",
		},
		{
			name:      "cat prefix (case 7)",
			randValue: 7,
			expected:  "Pia!<(=ｏ ‵-′)ノ☆ ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPiaPrefix(tt.randValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetOmen(t *testing.T) {
	tests := []struct {
		name      string
		randValue uint64
		expected  string
	}{
		{
			name:      "bad (value 0)",
			randValue: 0,
			expected:  "凶",
		},
		{
			name:      "bad (value 6)",
			randValue: 6,
			expected:  "凶",
		},
		{
			name:      "neutral (value 7)",
			randValue: 7,
			expected:  "",
		},
		{
			name:      "neutral (value 8)",
			randValue: 8,
			expected:  "",
		},
		{
			name:      "good (value 9)",
			randValue: 9,
			expected:  "吉",
		},
		{
			name:      "good (value 15)",
			randValue: 15,
			expected:  "吉",
		},
		{
			name:      "wraps around to bad (value 16)",
			randValue: 16,
			expected:  "凶",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOmen(tt.randValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMultiplier(t *testing.T) {
	tests := []struct {
		name      string
		randValue uint64
		expected  string
	}{
		{
			name:      "extremely small (value 0)",
			randValue: 0,
			expected:  "极小",
		},
		{
			name:      "super small (value 1)",
			randValue: 1,
			expected:  "超小",
		},
		{
			name:      "super small (value 10)",
			randValue: 10,
			expected:  "超小",
		},
		{
			name:      "ultra small (value 11)",
			randValue: 11,
			expected:  "特小",
		},
		{
			name:      "very small (value 56)",
			randValue: 56,
			expected:  "甚小",
		},
		{
			name:      "small (value 176)",
			randValue: 176,
			expected:  "小",
		},
		{
			name:      "neutral (value 386)",
			randValue: 386,
			expected:  "",
		},
		{
			name:      "large (value 638)",
			randValue: 638,
			expected:  "大",
		},
		{
			name:      "very large (value 848)",
			randValue: 848,
			expected:  "甚大",
		},
		{
			name:      "ultra large (value 968)",
			randValue: 968,
			expected:  "特大",
		},
		{
			name:      "super large (value 1013)",
			randValue: 1013,
			expected:  "超大",
		},
		{
			name:      "extremely large (value 1023)",
			randValue: 1023,
			expected:  "极大",
		},
		{
			name:      "wraps around to extremely small (value 1024)",
			randValue: 1024,
			expected:  "极小",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMultiplier(tt.randValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}
