package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host  string `default:"0.0.0.0"`
	Port  string `default:"8080"`
	Token string `require:"true"`
}

type Context struct {
	Rand  *rand.Rand
	Query *tgbotapi.InlineQuery
}

func newRand(seeds []uint64) *rand.Rand {
	var s uint64

	for _, v := range seeds {
		s ^= v
	}

	return rand.New(rand.NewSource(int64(s)))
}

func pia(ctx *Context) string {

	var pia string

	switch ctx.Rand.Uint64() % 8 {
	case 0:
		pia = "Pia!▼(ｏ ‵-′)ノ★"
	default:
		pia = "Pia!<(=ｏ ‵-′)ノ☆"
	}

	return fmt.Sprintf(
		"%s %s",
		pia,
		ctx.Query.Query,
	)

}

func divine(ctx *Context) string {
	var omen, mult, sign string

	o := ctx.Rand.Uint64() % 16

	switch {
	case 9 <= o:
		omen = "吉"
	case o < 7:
		omen = "凶"
	}

	if omen == "" {
		sign = "尚可"
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

		sign = mult + omen
	}

	return fmt.Sprintf(
		"所求事项: %s\n结果: %s\n",
		ctx.Query.Query,
		sign,
	)
}

func answerInline(q *tgbotapi.InlineQuery) tgbotapi.InlineConfig {
	h := sha256.New()

	if err := binary.Write(
		h,
		binary.LittleEndian,
		uint64(q.From.ID),
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
		[]byte(q.Query),
	); err != nil {
		log.Println(err)
	}

	r := bytes.NewReader(h.Sum(nil))

	// sha256 checksum is 256 bits long, equals to four 64bits integer.
	seeds := make([]uint64, 4)

	if err := binary.Read(r, binary.BigEndian, &seeds); err != nil {
		fmt.Println("binary.Read failed:", err)
	}

	ctx := Context{
		Rand:  newRand(seeds),
		Query: q,
	}

	var res []interface{} = make([]interface{}, 2)

	res[0] = tgbotapi.NewInlineQueryResultArticleMarkdown(
		"divine",
		"求签",
		divine(&ctx),
	)

	res[1] = tgbotapi.NewInlineQueryResultArticleMarkdown(
		"pia",
		"Pia",
		pia(&ctx),
	)

	ans := tgbotapi.InlineConfig{
		InlineQueryID: q.ID,
		Results:       res,
	}
	return ans
}

func main() {
	var conf Config

	envconfig.MustProcess("", &conf)

	bot, err := tgbotapi.NewBotAPI(conf.Token)
	if err != nil {
		log.Fatal(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		tm := time.Unix(int64(info.LastErrorDate), 0)
		log.Printf("Last error: %s %s", tm, info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/")

	go http.ListenAndServe(net.JoinHostPort(conf.Host, conf.Port), nil)

	for update := range updates {
		if update.InlineQuery != nil {

			ans := answerInline(update.InlineQuery)

			resp, err := bot.AnswerInlineQuery(ans)

			if err != nil {
				log.Println("Error: ", err)
			} else if !resp.Ok {
				log.Println(
					"Error: ",
					resp.ErrorCode,
					resp.Description,
				)
			}
		}
	}
}
