package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
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
	Host  string `default:"127.0.0.1"`
	Port  string `default:"8080"`
	Token string `require:"true"`
}

func divine(input []byte) string {
	var sum int

	// modulo is distributive
	// n_1 + n_2 + ... + n_i mod k
	//     = (n_1 mod k + n_2 mod k + ... + n_i mod k) mod k
	//
	// Should work as long as there are less than 2^(63 - 8) elements in a
	// slice.
	for _, v := range input {
		sum += int(v)
		sum %= 256
	}

	var res string

	// Binomial coefficient choose(8, 0:8)
	// TODO internationalization
	switch {
	case 254 < sum:
		res = "超大吉"
	case 247 < sum && sum <= 247:
		res = "大吉"
	case 219 < sum && sum <= 247:
		res = "吉"
	case 163 < sum && sum <= 219:
		res = "小吉"
	case 93 < sum && sum <= 163:
		res = "尚可"
	case 37 < sum && sum <= 93:
		res = "小凶"
	case 9 < sum && sum <= 37:
		res = "凶"
	case 2 < sum && sum <= 9:
		res = "大凶"
	case sum <= 2:
		res = "超大凶"
	default:
		res = "???"
	}

	return res
}

func answerInline(q *tgbotapi.InlineQuery) tgbotapi.InlineConfig {
	var b bytes.Buffer

	binary.Write(
		&b,
		binary.LittleEndian,
		uint64(q.From.ID),
	)

	t := time.Now().Truncate(30 * time.Minute).Unix()
	rand.Seed(t)

	binary.Write(
		&b,
		binary.LittleEndian,
		t,
	)

	binary.Write(
		&b,
		binary.LittleEndian,
		rand.Uint64(),
	)

	binary.Write(
		&b,
		binary.LittleEndian,
		q.Query,
	)

	chksum := sha256.Sum256(b.Bytes())

	id := hex.EncodeToString(chksum[:])

	var res []interface{} = make([]interface{}, 1)

	res[0] = tgbotapi.NewInlineQueryResultArticleHTML(
		id,
		"求签",
		fmt.Sprintf("所求事项: %s\n结果: %s\n", q.Query, divine(chksum[:])),
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
