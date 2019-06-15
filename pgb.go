package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	host, port string
	token      string
)

func answerInline(q *tgbotapi.InlineQuery) tgbotapi.InlineConfig {
	var b bytes.Buffer

	binary.Write(
		&b,
		binary.LittleEndian,
		uint64(q.From.ID),
	)

	binary.Write(
		&b,
		binary.LittleEndian,
		time.Now().Unix(),
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
		"大凶",
	)

	ans := tgbotapi.InlineConfig{
		InlineQueryID: q.ID,
		Results:       res,
	}
	return ans
}

func main() {
	flag.StringVar(
		&host,
		"h",
		"127.0.0.1",
		"IP address of the bot to bind to, default to 127.0.0.1.",
	)

	flag.StringVar(
		&port,
		"p",
		"8080",
		"port of the bot, default to 8080",
	)

	flag.StringVar(
		&token,
		"t",
		"",
		"bot token",
	)
	flag.Parse()

	if token == "" {
		log.Fatalln("No token set.")
	}

	bot, err := tgbotapi.NewBotAPI(token)
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

	go http.ListenAndServe(net.JoinHostPort(host, port), nil)

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
