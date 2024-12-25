package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
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
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host  string `default:"0.0.0.0"`
	Port  string `default:"8080"`
	Token string `require:"true"`
}

type UpdateContext struct {
	Rand  *rand.Rand
	Query *string
}

type builder struct {
	strings.Builder
}

func (b *builder) WriteStrings(strs ...string) {
	for _, s := range strs {
		b.WriteString(s)
	}
}

func newRand(seeds []uint64) *rand.Rand {
	var s uint64

	for _, v := range seeds {
		s ^= v
	}

	return rand.New(rand.NewSource(int64(s)))
}

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

	envconfig.MustProcess("", &conf)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

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

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
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
		fmt.Println("binary.Read failed:", err)
	}

	rctx := UpdateContext{
		Rand:  newRand(seeds),
		Query: &update.InlineQuery.Query,
	}

	results := []models.InlineQueryResult{
		&models.InlineQueryResultArticle{ID: "divine", Title: "求签", InputMessageContent: &models.InputTextMessageContent{MessageText: divine(&rctx)}},
		&models.InlineQueryResultArticle{ID: "2", Title: "Pia", InputMessageContent: &models.InputTextMessageContent{MessageText: pia(&rctx)}},
	}

	b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
		InlineQueryID: update.InlineQuery.ID,
		Results:       results,
	})
}
