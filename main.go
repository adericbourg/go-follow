package main

import (
	"fmt"
	"os"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"log"
	"strings"
	"math/rand"
	"time"
	"mvdan.cc/xurls"
	"encoding/xml"
	"net/http"
	"io/ioutil"
	"bufio"
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

	currentUser, err := getCurrentUser(client)
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch current user (%s)", err))
	}

	context := Context{
		Client:      client,
		User:        currentUser,
		LastRetweet: time.Now().AddDate(-1, 0, 0),
		LastComment: time.Now().AddDate(-1, 0, 0),
		LastLink:    time.Now().AddDate(-1, 0, 0),
		LastFollow:  time.Now().AddDate(-1, 0, 0),
		Stats:       Stats{},
		Rates: Rates{
			Follow:   CreateRateLimiter(2, time.Hour),
			Favorite: CreateRateLimiter(1, time.Minute*5),
			Status:   CreateRateLimiter(1, time.Minute*5),
			Retweet:  CreateRateLimiter(1, time.Minute*5),
			Unfollow: CreateRateLimiter(3, time.Hour),
		},
	}

	go scheduleEvery(time.Minute, func(t time.Time) {
		logStats(&context)
	})

	go scheduleEvery(time.Hour, func(t time.Time) {
		pruneFriends(&context)
	})

	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		handleTweet(tweet, &context)
	}

	log.Println("Starting stream processing")
	demux.HandleChan(stream.Messages)

	fmt.Printf("Press [enter] to quit")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	log.Printf("Received '%s'. Exiting)", input)
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
	User        *twitter.User
	LastRetweet time.Time
	LastComment time.Time
	LastLink    time.Time
	LastFollow  time.Time
	Stats       Stats
	Rates       Rates
}

type Stats struct {
	Retweets int32
	Comments int32
	Favorite int32
	Follow   int32
	Links    int32
	Ignore   int32
}

type Rates struct {
	Retweet  *RateLimiter
	Status   *RateLimiter
	Favorite *RateLimiter
	Follow   *RateLimiter
	Unfollow *RateLimiter
}

func logStats(context *Context) {
	log.Printf("Stats %+v\n", context.Stats)

	log.Println(" == Rates ==")
	logRate("Retweets ", context.Rates.Retweet)
	logRate("Status   ", context.Rates.Status)
	logRate("Favorite ", context.Rates.Favorite)
	logRate("Follow   ", context.Rates.Follow)
	logRate("Unfollow ", context.Rates.Unfollow)

	log.Println()
}

func logRate(label string, rateLimiter *RateLimiter) {
	log.Printf("  %s: %d / %d (%s)\n", label, rateLimiter.reservoir.Sum(), rateLimiter.threshold, rateLimiter.reservoir.TimeWindow)
}

func getCurrentUser(client *twitter.Client) (*twitter.User, error) {
	currentUser, _, err := client.Accounts.VerifyCredentials(&twitter.AccountVerifyParams{})

	return currentUser, err
}

var TweetHandlers = []func(*twitter.Tweet, *Context){
	retweet,
	comment,
	favorite,
	follow,
}

func handleTweet(tweet *twitter.Tweet, context *Context) {
	if tweet.User.ID == context.User.ID {
		return
	}
	if strings.HasPrefix(tweet.Text, "RT ") {
		context.Stats.Ignore += 1
		return
	}
	if tweet.InReplyToStatusID != 0 {
		context.Stats.Ignore += 1
		return
	}

	tweetUrls := getUrls(tweet)
	if len(tweetUrls) > 1 {
		postLink(tweetUrls[0], context)
		return
	}

	handler := TweetHandlers[rand.Intn(len(TweetHandlers))]
	handler(tweet, context)
}

func retweet(tweet *twitter.Tweet, context *Context) {
	context.Rates.Retweet.Submit(func() {
		context.Stats.Retweets += 1
		context.Client.Statuses.Retweet(tweet.ID, &twitter.StatusRetweetParams{
			ID: tweet.ID,
		})
		context.LastRetweet = time.Now()
	})
}

var Comments = []string{
	"Thanks!",
	"Thank you for this :)",
	"Interesting. Thanks for sharing!",
	"👍",
	"Great!",
	"Interesting point. Could you elaborate?",
	"How can I get more details about this?",
	"nice :)",
	"Can you tell us more about that?",
}

func comment(tweet *twitter.Tweet, context *Context) {
	context.Rates.Status.Submit(func() {
		context.Stats.Comments += 1
		comment := fmt.Sprintf("@%s %s", tweet.User.ScreenName, Comments[ rand.Intn(len(Comments))])
		context.Client.Statuses.Update(comment, &twitter.StatusUpdateParams{
			InReplyToStatusID: tweet.ID,
		})
		context.LastComment = time.Now()
	})
}

func favorite(tweet *twitter.Tweet, context *Context) {
	context.Rates.Favorite.Submit(func() {
		context.Stats.Favorite += 1
		context.Client.Favorites.Create(&twitter.FavoriteCreateParams{
			ID: tweet.ID,
		})
	})
}

func follow(tweet *twitter.Tweet, context *Context) {
	context.Rates.Follow.Submit(func() {
		context.Stats.Follow += 1
		context.Client.Friendships.Create(&twitter.FriendshipCreateParams{
			UserID:     tweet.User.ID,
			ScreenName: tweet.User.ScreenName,
		})
		context.LastFollow = time.Now()
	})
}

func getUrls(tweet *twitter.Tweet) []string {
	return xurls.Relaxed().FindAllString(tweet.Text, -1)
}

func postLink(url string, context *Context) {
	context.Rates.Status.Submit(func() {
		title, err := getTitle(url)

		if err != nil {
			return
		}

		context.Stats.Links += 1
		status := fmt.Sprintf("%s\n%s #blockchain", title, url)
		context.Client.Statuses.Update(status, &twitter.StatusUpdateParams{})
		context.LastLink = time.Now()
	})
}

type Html struct {
	Head Head `xml:"head"`
}

type Head struct {
	Title string `xml:"title"`
}

func getTitle(url string) (string, error) {
	response, err := http.Get(url)

	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	var html Html
	xml.Unmarshal([]byte(body), &html)

	return html.Head.Title, nil
}

func pruneFriends(context *Context) {
	pruneCountTarget, err := computePruneTarget(context)

	if err != nil {
		log.Printf("Unable to compute prune target. Abort.")
		return
	}

	pruned := 0
	var cursor int64 = -1
	for {
		friends, _, _ := context.Client.Friends.IDs(&twitter.FriendIDParams{
			Cursor: cursor,
		})
		cursor = friends.NextCursor

		friendIds := friends.IDs
		shuffle(friendIds)

		for _, friendId := range friendIds {
			context.Rates.Status.Submit(func() {
				context.Client.Friendships.Destroy(&twitter.FriendshipDestroyParams{
					UserID: friendId,
				})
			})
			pruned += 1
			if pruned >= pruneCountTarget {
				goto TheEnd
			}
		}

		if cursor == 0 {
			goto TheEnd
		}
	}
TheEnd:
	log.Printf("Pruned %d friends (out of a %d target)", pruned, pruneCountTarget)

}

func computePruneTarget(context *Context) (int, error) {
	user, err := getCurrentUser(context.Client)

	if err != nil {
		return 0, err
	}

	followerCount := user.FollowersCount
	friendsCount := user.FriendsCount

	const pruneRatio float32 = 1.1

	pruneCountTarget := int(float32(friendsCount-followerCount) * pruneRatio)

	return pruneCountTarget, nil
}

func shuffle(slice []int64) {
	for i := range slice {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}
