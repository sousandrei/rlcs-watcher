package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/araddon/dateparse"
	"github.com/go-redis/redis"
	tb "gopkg.in/tucnak/telebot.v2"
)

func logFatal(msg string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", err, msg)
	}
}

const REDIS_KEY = "latest-rlcs"
const BASE_URL = "https://esports.rocketleague.com"

func createBot() *tb.Bot {
	botToken := os.Getenv("BOT_TOKEN")

	bot, err := tb.NewBot(tb.Settings{
		Token:  botToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	logFatal("Cannot create telegram bot", err)

	bot.Handle(tb.OnText, handleMe)

	return bot
}

func handleMe(m *tb.Message) {
	log.Println(m.Sender.ID, m.Text)
}

func main() {

	bot := createBot()
	log.Println("Bot created")

	go bot.Start()
	log.Println("Bot started")

	chatId, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	logFatal("CHAT_ID not present", err)

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		log.Fatalf("REDIS_HOST not present")
	}

	redisPass := os.Getenv("REDIS_PASSWORD")
	if redisPass == "" {
		log.Fatalf("REDIS_PASSWORD not present")
	}

	log.Println("Starting redis client")
	db := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPass,
		DB:       0,
	})

	log.Println("Starting loop")
	for {
		time.Sleep(1 * time.Minute)
		log.Println("Processing")

		res, err := http.Get(BASE_URL + "/news")

		if err != nil {
			log.Println("Cannot get news: ", err)
			continue
		}

		if res.StatusCode != 200 {
			log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		res.Body.Close()
		logFatal("Error parsing body", err)

		latest_news := doc.Find("#article-content > div > div:nth-child(1) > a > div > div > p")
		link := doc.Find("#article-content > div > div:nth-child(1) > a").AttrOr("href", "")

		latest_date, err := dateparse.ParseStrict(strings.Split(latest_news.Text(), " - ")[0])
		logFatal("Cannot parse date", err)

		latest_date_unix := latest_date.Unix()

		value, err := db.Get(REDIS_KEY).Result()

		var latest_stored_value int

		if err != nil {
			latest_stored_value = 0
		} else {
			latest_stored_value, err = strconv.Atoi(value)
			if err != nil {
				log.Println("Failed to parse latest redis key")
			}
		}

		if latest_stored_value < int(latest_date_unix) {
			err := db.Set(REDIS_KEY, latest_date_unix, 0).Err()
			if err != nil {
				log.Println("Failed to set redis key on new entry")
			}

			log.Println("sending message")
			bot.Send(&tb.Chat{ID: chatId}, BASE_URL+link)
		}

	}

}
