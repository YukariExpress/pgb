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

// Config represents the configuration settings for the application.
// It includes the following fields:
//   - Host: The hostname or IP address where the application will run.
//     It is set via the "HOST" environment variable and defaults to "0.0.0.0".
//   - Port: The port number on which the application will listen.
//     It is set via the "PORT" environment variable and defaults to "8080".
//   - Token: A required authentication token for the application.
//     It is set via the "TOKEN" environment variable.
type Config struct {
	Host  string `env:"HOST, default=0.0.0.0"`
	Port  string `env:"PORT, default=8080"`
	Token string `env:"TOKEN, required"`
}

// UpdateContext holds the context for an update operation.
// It includes a random number generator, a query string, and a locale.
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

// WriteStrings writes multiple strings to the builder.
// It takes a variadic parameter of strings and writes each one
// sequentially using the WriteString method.
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

// pia generates a string based on the provided UpdateContext.
// It randomly selects one of two possible pia (slap) actions and appends the query from the context.
//
// Parameters:
//   - ctx: A pointer to an UpdateContext containing the query and random number generator.
//
// Returns:
//
//	A string that includes a randomly selected pia prefix and the query from the context.
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

// divine generates a divination result based on the provided UpdateContext.
// It constructs a string that includes the query and the result of the divination.
// The result is determined by generating random numbers and mapping them to specific outcomes.
// The outcomes are categorized as "吉" (good) or "凶" (bad) with varying degrees of intensity.
//
// Parameters:
// - ctx: A pointer to an UpdateContext which contains the query and a random number generator.
//
// Returns:
// - A string representing the divination result.
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
	}

	if b, err := bot.New(conf.Token, opts...); nil != err {
		panic(err)
	} else {
		go b.StartWebhook(ctx)

		http.ListenAndServe(net.JoinHostPort(conf.Host, conf.Port), b.WebhookHandler())
	}
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
	locale := update.InlineQuery.From.LanguageCode

	if update.InlineQuery == nil {
		return
	}

	h := sha256.New()

	if err := binary.Write(
		h,
		binary.LittleEndian,
		uint64(update.InlineQuery.From.ID),
	); err != nil {
		log.Println(err)
	}

	if err := binary.Write(
		h,
		binary.LittleEndian,
		time.Now().Truncate(30*time.Minute).Unix(),
	); err != nil {
		log.Println(err)
	}

	if err := binary.Write(
		h,
		binary.LittleEndian,
		[]byte(update.InlineQuery.Query),
	); err != nil {
		log.Println(err)
	}

	r := bytes.NewReader(h.Sum(nil))

	// sha256 checksum is 256 bits long, equals to four 64bits integer.
	seeds := make([]uint64, 4)

	if err := binary.Read(r, binary.BigEndian, &seeds); err != nil {
		log.Println("binary.Read failed:", err)
	}

	rctx := UpdateContext{
		Rand:   newRand(seeds),
		Query:  &update.InlineQuery.Query,
		Locale: &locale,
	}

	var divineTitle, piaTitle string

	switch locale {
	case "zh":
		divineTitle = "求签"
		piaTitle = "Pia"
	default:
		divineTitle = "Divination"
		piaTitle = "Pia"
	}

	results := []models.InlineQueryResult{
		&models.InlineQueryResultArticle{
			ID:    "divine",
			Title: divineTitle,
			InputMessageContent: &models.InputTextMessageContent{
				MessageText: divine(&rctx),
			},
		},
		&models.InlineQueryResultArticle{
			ID:    "pia",
			Title: piaTitle,
			InputMessageContent: &models.InputTextMessageContent{
				MessageText: pia(&rctx),
			},
		},
	}

	b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
		InlineQueryID: update.InlineQuery.ID,
		Results:       results,
	})
}
