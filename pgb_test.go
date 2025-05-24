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
