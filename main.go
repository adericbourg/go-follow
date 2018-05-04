package main

import (
	"fmt"
	"os"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"os/signal"
	"syscall"
	"log"
	"strings"
	"math/rand"
	"time"
)

func main() {
	consumerKey := getenv("TWITTER_CONSUMER_KEY")
	consumerSecret := getenv("TWITTER_CONSUMER_SECRET")
	accessToken := getenv("TWITTER_ACCESS_TOKEN")
	accessTokenSecret := getenv("TWITTER_ACCESS_TOKEN_SECRET")

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)

	params := &twitter.StreamFilterParams{
		Track:         []string{"goland"},
		StallWarnings: twitter.Bool(true),
	}

	stream, err := client.Streams.Filter(params)
	if err != nil {
		panic("Failed to build stream")
	}

	demux := twitter.NewSwitchDemux()

	latestRetweet := time.Now().AddDate(-1, 0, 0)
	demux.Tweet = func(tweet *twitter.Tweet) {
		if strings.HasPrefix(tweet.Text, "RT ") {
			return
		}

		duration := time.Since(latestRetweet)
		// 10min + random between 0 and 20min
		if duration.Minutes() >= (10 + rand.Float64()*20) {
			fmt.Printf("Retweet: %s\n", tweet.Text)
			client.Statuses.Retweet(tweet.ID, &twitter.StatusRetweetParams{
				ID: tweet.ID,
			})
			latestRetweet = time.Now()
		}
	}

	demux.HandleChan(stream.Messages)

	// Wait for SIGINT and SIGTERM
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	fmt.Println("Exiting")
	stream.Stop()
}

func getenv(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		panic(fmt.Sprintf("Missing environment variable: %s", key))
	}
	return value
}
