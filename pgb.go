package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	host, port string
	token      string
)

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

			var b bytes.Buffer

			binary.Write(
				&b,
				binary.LittleEndian,
				update.InlineQuery.From.ID,
			)

			binary.Write(
				&b,
				binary.LittleEndian,
				update.InlineQuery.Query,
			)

			chksum := sha256.Sum256(b.Bytes())

			id := hex.EncodeToString(chksum[:])

			res := []tgbotapi.InlineQueryResultArticle{
				tgbotapi.InlineQueryResultArticle{
					ID:   id,
					Type: "article",
					InputMessageContent: tgbotapi.InputTextMessageContent{
						Text:      "大凶",
						ParseMode: "HTML",
					},
					Title: "求签",
				},
			}

			var resIterfaces []interface{} = make([]interface{}, len(res))
			for i, d := range res {
				resIterfaces[i] = d
			}

			ans := tgbotapi.InlineConfig{
				InlineQueryID: update.InlineQuery.ID,
				Results:       resIterfaces,
			}

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
