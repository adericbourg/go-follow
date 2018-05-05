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
		Track:         []string{"blockchain"},
		StallWarnings: twitter.Bool(true),
	}

	stream, err := client.Streams.Filter(params)
	if err != nil {
		panic("Failed to build stream")
	}

	context := Context{
		Client:      client,
		LastRetweet: time.Now().AddDate(-1, 0, 0),
		LastComment: time.Now().AddDate(-1, 0, 0),
		Stats: Stats{
			Comments: 0,
			Favorite: 0,
			Follow:   0,
			Ignore:   0,
			Retweets: 0,
		},
	}

	go scheduleEvery(30*time.Second, func(t time.Time) {
		logStats(&context)
	})

	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		handleTweet(tweet, &context)
	}

	log.Println("Starting stream processing")
	demux.HandleChan(stream.Messages)

	// Wait for SIGINT and SIGTERM
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Exiting")
	stream.Stop()
}

func getenv(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		panic(fmt.Sprintf("Missing environment variable: %s", key))
	}
	return value
}

func scheduleEvery(d time.Duration, f func(time.Time)) {
	f(time.Now())
	for x := range time.Tick(d) {
		f(x)
	}
}

type Context struct {
	Client      *twitter.Client
	LastRetweet time.Time
	LastComment time.Time
	Stats       Stats
}

type Stats struct {
	Retweets int32
	Comments int32
	Favorite int32
	Follow   int32
	Ignore   int32
}

func logStats(context *Context) {
	log.Printf(
		"Stats { Retweets: %d, Comments: %d, Favorite: %d, Follow: %d, Ignore: %d }",
		context.Stats.Retweets, context.Stats.Comments, context.Stats.Favorite, context.Stats.Follow, context.Stats.Ignore)
}

func handleTweet(tweet *twitter.Tweet, context *Context) {
	if strings.HasPrefix(tweet.Text, "RT ") {
		context.Stats.Ignore += 1
		return
	}
	if tweet.InReplyToStatusID == 0 {
		context.Stats.Ignore += 1
		return
	}

	processId := rand.Intn(4)
	switch  processId {
	case 0:
		retweet(tweet, context)
	case 1:
		comment(tweet, context)
	case 2:
		favorite(tweet, context)
	case 3:
		follow(tweet, context)
	default:
		favorite(tweet, context)
	}
}

func retweet(tweet *twitter.Tweet, context *Context) {
	duration := time.Since(context.LastRetweet)
	// 10min + random between 0 and 20min
	if duration.Minutes() >= (10 + rand.Float64()*20) {
		context.Stats.Retweets += 1
		context.Client.Statuses.Retweet(tweet.ID, &twitter.StatusRetweetParams{
			ID: tweet.ID,
		})
		context.LastRetweet = time.Now()
	}
}

var Comments = []string{
	"Thanks!",
	"Thank you for this :)",
	"Interesting. Thanks for sharing!",
	"👍",
}

func comment(tweet *twitter.Tweet, context *Context) {
	duration := time.Since(context.LastComment)

	// 10min + random between 0 and 5min
	if duration.Minutes() >= (10 + rand.Float64()*5) {
		context.Stats.Comments += 1
		comment := Comments[ rand.Intn(len(Comments))]
		context.Client.Statuses.Update(comment, &twitter.StatusUpdateParams{
			Status:            comment,
			InReplyToStatusID: tweet.ID,
		})
		context.LastComment = time.Now()
	}
}

func favorite(tweet *twitter.Tweet, context *Context) {
	context.Stats.Favorite += 1
	context.Client.Favorites.Create(&twitter.FavoriteCreateParams{
		ID: tweet.ID,
	})
}

func follow(tweet *twitter.Tweet, context *Context) {
	context.Stats.Follow += 1
	context.Client.Friendships.Create(&twitter.FriendshipCreateParams{
		UserID:     tweet.User.ID,
		ScreenName: tweet.User.ScreenName,
	})
}
