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

func newRand(seeds []uint64) *rand.Rand {
	var s uint64

	for _, v := range seeds {
		s ^= v
	}

	return rand.New(rand.NewSource(int64(s)))
}

func pia(r *rand.Rand) string {
	switch r.Uint64() % 8 {
	case 0:
		return "Pia!▼(ｏ ‵-′)ノ★"
	default:
		return "Pia!<(=ｏ ‵-′)ノ☆"
	}
}

func divine(r *rand.Rand) string {
	var omen, mult string

	switch r.Uint64() % 16 {
	case 10, 11, 12, 13, 14, 15:
		omen = "吉"
	case 0, 1, 2, 3, 4, 5, 6:
		omen = "凶"
	}

	if omen == "" {
		return "尚可"
	}

	m := r.Uint64() % 1024

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

	return mult + omen
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

	ra := newRand(seeds)

	var res []interface{} = make([]interface{}, 2)

	res[0] = tgbotapi.NewInlineQueryResultArticleMarkdown(
		"divine",
		"求签",
		fmt.Sprintf(
			"所求事项: %s\n结果: %s\n",
			q.Query,
			divine(ra),
		),
	)

	res[1] = tgbotapi.NewInlineQueryResultArticleMarkdown(
		"pia",
		"Pia",
		fmt.Sprintf(
			"%s %s",
			pia(ra),
			q.Query,
		),
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
