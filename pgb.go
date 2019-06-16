package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc64"
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

func divine(in int64) string {

	r := rand.New(rand.NewSource(in))

	var omen, mult string

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

	switch r.Uint64() % 16 {
	case 10, 11, 12, 13, 14, 15:
		omen = "吉"
	case 0, 1, 2, 3, 4, 5, 6:
		omen = "凶"
	}

	if omen == "" {
		return "尚可"
	}
	return mult + omen
}

func answerInline(q *tgbotapi.InlineQuery) tgbotapi.InlineConfig {
	h := crc64.New(crc64.MakeTable(crc64.ISO))

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

	chksum := h.Sum64()

	var res []interface{} = make([]interface{}, 1)

	res[0] = tgbotapi.NewInlineQueryResultArticleHTML(
		fmt.Sprintf("%x", chksum),
		"求签",
		fmt.Sprintf(
			"所求事项: %s\n结果: %s\n",
			q.Query,
			divine(int64(chksum)),
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
