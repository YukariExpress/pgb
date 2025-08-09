// PGB: Pythia Gata Bot
// Copyright (C) 2019-2024  Yishen Miao
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sethvargo/go-envconfig"
)

// Config represents the configuration settings for the application. It includes
// the following fields:
//
//   - Host: The hostname or IP address where the application will run. It is
//     set via the "HOST" environment variable and defaults to "0.0.0.0".
//   - Port: The port number on which the application will listen. It is set via
//     the "PORT" environment variable and defaults to "8080".
//   - Token: A required authentication token for the application. It is set via
//     the "TOKEN" environment variable.
type Config struct {
	Debug   bool          `env:"DEBUG, default=false"`
	Host    string        `env:"HOST, default=0.0.0.0"`
	Port    string        `env:"PORT, default=8080"`
	Timeout time.Duration `env:"TIMEOUT, default=5s"`
	Token   string        `env:"TOKEN, required"`
}

// UpdateContext holds the context for an update operation. It includes a random
// number generator, a query string, and a locale.
//
// Fields:
// - Rand: A pointer to a rand.Rand instance used for generating random numbers.
// - Query: A pointer to a string representing the query to be executed.
// - Locale: A string representing the locale of the user.
type UpdateContext struct {
	Rand   *rand.Rand
	Query  *string
	Locale *string
}

// builder is a custom type that embeds strings.Builder to provide additional
// functionality or methods specific to the application. It inherits all the
// methods of strings.Builder and can be used in the same way.
type builder struct {
	strings.Builder
}

// WriteStrings writes multiple strings to the builder. It takes a variadic
// parameter of strings and writes each one sequentially using the WriteString
// method.
//
// Parameters:
//
//	strs - a variadic list of strings to be written to the builder.
func (b *builder) WriteStrings(strs ...string) {
	for _, s := range strs {
		b.WriteString(s)
	}
}

// newRand creates a new instance of rand.Rand seeded with the XOR combination
// of the provided seeds. It takes a slice of uint64 seeds, combines them using
// the XOR operation, and returns a pointer to a new rand.Rand object seeded
// with the resulting value.
//
// Parameters:
//   - seeds: A slice of uint64 values used to seed the random number generator.
//
// Returns:
//   - A pointer to a new rand.Rand object seeded with the combined seed value.
func newRand(seeds []uint64) *rand.Rand {
	var s uint64

	for _, v := range seeds {
		s ^= v
	}

	return rand.New(rand.NewSource(int64(s)))
}

// pia generates a string based on the provided UpdateContext.  It randomly
// selects one of two possible pia (slap) actions performed by either a dog or a
// cat and appends the query from the context. There is a 1 in 8 chance to
// summon a dog and a 7 in 8 chance to summon a cat.
//
// Parameters:
//   - ctx: A pointer to an UpdateContext containing the query and random number
//     generator.
//
// Returns:
//   - A string that includes a randomly selected pia prefix and the query from
//     the context.
func pia(ctx *UpdateContext) string {
	var b builder

	switch ctx.Rand.Uint64() % 8 {
	case 0:
		b.WriteString("Pia!▼(ｏ ‵-′)ノ★ ")
	default:
		b.WriteString("Pia!<(=ｏ ‵-′)ノ☆ ")
	}

	b.WriteString(*ctx.Query)

	return b.String()
}

// divine generates a divination result based on the provided UpdateContext. It
// constructs a string that includes the query and the result of the divination.
// The result is determined by generating random numbers and mapping them to
// specific outcomes. The outcomes are categorized as "吉" (good) or "凶" (bad)
// with varying degrees of intensity.
//
// Parameters:
//   - ctx: A pointer to an UpdateContext which contains the query and a random
//     number generator.
//
// Returns:
//   - A string representing the divination result.
func divine(ctx *UpdateContext) string {
	var omen, mult string
	var b builder

	b.WriteStrings("所求事项: ", *ctx.Query, "\n结果: ")

	o := ctx.Rand.Uint64() % 16

	switch {
	case 9 <= o:
		omen = "吉"
	case o < 7:
		omen = "凶"
	}

	if omen == "" {
		b.WriteString("尚可")
	} else {

		m := ctx.Rand.Uint64() % 1024

		switch {
		case m < 1:
			mult = "极小"
		case 1 <= m && m < 11:
			mult = "超小"
		case 11 <= m && m < 56:
			mult = "特小"
		case 56 <= m && m < 176:
			mult = "甚小"
		case 176 <= m && m < 386:
			mult = "小"
		case 386 <= m && m < 638:
			mult = ""
		case 638 <= m && m < 848:
			mult = "大"
		case 848 <= m && m < 968:
			mult = "甚大"
		case 968 <= m && m < 1013:
			mult = "特大"
		case 1013 <= m && m < 1023:
			mult = "超大"
		case 1023 <= m:
			mult = "极大"
		}

		b.WriteStrings(mult, omen)
	}

	return b.String()
}

func main() {
	var conf Config

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := envconfig.Process(ctx, &conf); err != nil {
		log.Fatal(err)
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
		bot.WithCheckInitTimeout(conf.Timeout),
	}

	if conf.Debug {
		opts = append(opts, bot.WithDebug())
	}

	if b, err := bot.New(conf.Token, opts...); nil != err {
		panic(err)
	} else {
		go b.StartWebhook(ctx)

		http.ListenAndServe(
			net.JoinHostPort(conf.Host, conf.Port),
			b.WebhookHandler(),
		)
	}
}

// getLocaleTitles returns the localized titles for the divine and pia results
// based on the locale string.
//
// Parameters:
//   - locale: the user's language code (e.g., "zh", "en").
//
// Returns:
//   - divineTitle: the localized title for the divination result.
//   - piaTitle: the localized title for the pia result.
func getLocaleTitles(locale string) (string, string) {
	switch locale {
	case "zh":
		return "求签", "Pia"
	default:
		return "Divination", "Pia"
	}
}

// getUserID extracts the user ID as uint64 from a models.User pointer. Returns
// 0 if the user is nil.
//
// Parameters:
//   - user: pointer to a models.User struct.
//
// Returns:
//   - userID: the user's ID as uint64, or 0 if user is nil.
func getUserID(user *models.User) uint64 {
	if user != nil {
		return uint64(user.ID)
	}
	return 0
}

// getUserLocale extracts the language code from a models.User pointer. Returns
// "zh" if the user is nil or the language code is empty.
//
// Parameters:
//   - user: pointer to a models.User struct.
//
// Returns:
//   - locale: the user's language code, or "zh" as default.
func getUserLocale(user *models.User) string {
	if user != nil && user.LanguageCode != "" {
		return user.LanguageCode
	}
	return "zh"
}

// buildUpdateContext creates an UpdateContext for a given user ID, query, and
// locale. It generates a deterministic random number generator seeded by the
// user ID, time, and query.
//
// Parameters:
//   - userID: the user's ID as uint64.
//   - queryText: the query string.
//   - locale: the user's locale string.
//
// Returns:
//   - pointer to an UpdateContext struct.
func buildUpdateContext(userID uint64, queryText, locale string) *UpdateContext {
	h := sha256.New()
	_ = binary.Write(h, binary.LittleEndian, userID)
	_ = binary.Write(h, binary.LittleEndian, time.Now().Truncate(30*time.Minute).Unix())
	_ = binary.Write(h, binary.LittleEndian, []byte(queryText))
	r := bytes.NewReader(h.Sum(nil))
	seeds := make([]uint64, 4)
	_ = binary.Read(r, binary.BigEndian, &seeds)
	return &UpdateContext{
		Rand:   newRand(seeds),
		Query:  &queryText,
		Locale: &locale,
	}
}

// buildInlineQueryResults generates the inline query results for a given user
// and query text. It determines the locale and user ID, builds the
// UpdateContext, and returns the results.
//
// Parameters:
//   - user: pointer to a models.User struct (may be nil).
//   - queryText: the query string.
//
// Returns:
//   - slice of models.InlineQueryResult containing the divine and pia articles.
func buildInlineQueryResults(user *models.User, queryText string) []models.InlineQueryResult {
	locale := getUserLocale(user)
	userID := getUserID(user)

	rctx := buildUpdateContext(userID, queryText, locale)
	divineTitle, piaTitle := getLocaleTitles(locale)

	results := []models.InlineQueryResult{
		&models.InlineQueryResultArticle{
			ID:    "divine",
			Title: divineTitle,
			InputMessageContent: &models.InputTextMessageContent{
				MessageText: divine(rctx),
			},
		},
		&models.InlineQueryResultArticle{
			ID:    "pia",
			Title: piaTitle,
			InputMessageContent: &models.InputTextMessageContent{
				MessageText: pia(rctx),
			},
		},
	}
	return results
}

// handler processes an incoming inline query from a bot and generates a
// response.  It uses the query details and current time to create a unique
// context for the query, then generates a set of inline query results based on
// this context and sends them back.
//
// Parameters:
//   - ctx: The context for the request, used for cancellation and deadlines.
//   - b: The bot instance handling the request.
//   - update: The update containing the inline query to be processed.
//
// The function performs the following steps:
//  1. Checks if the update contains an inline query. If not, it returns
//     immediately.
//  2. Creates a SHA-256 hash based on the user's ID, current time truncated to
//     30 minutes, and the query text.
//  3. Uses the hash to seed a random number generator.
//  4. Creates an UpdateContext with the random number generator and query text.
//  5. Generates a set of inline query results using the UpdateContext.
//  6. Sends the generated results back to the bot as a response to the inline
//     query.
func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.InlineQuery == nil {
		return
	}
	user := update.InlineQuery.From
	queryText := update.InlineQuery.Query
	results := buildInlineQueryResults(user, queryText)
	b.AnswerInlineQuery(
		ctx,
		&bot.AnswerInlineQueryParams{
			InlineQueryID: update.InlineQuery.ID,
			Results:       results,
		},
	)
}
