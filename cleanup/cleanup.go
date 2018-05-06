package main

import (
	"github.com/dghubble/oauth1"
	"github.com/dghubble/go-twitter/twitter"
	"log"
	"os"
	"fmt"
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

	cleanupTweets(client)
	cleanupFriends(client)
	cleanupFavorites(client)
}

func getenv(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		panic(fmt.Sprintf("Missing environment variable: %s", key))
	}
	return value
}

func cleanupTweets(client *twitter.Client) {
	removed := 0
	for {
		tweets, _, _ := client.Timelines.UserTimeline(&twitter.UserTimelineParams{})
		if len(tweets) == 0 {
			return
		}

		for _, tweet := range tweets {
			client.Statuses.Destroy(tweet.ID, &twitter.StatusDestroyParams{})
			removed += 1
		}
	}
	log.Printf("Remmoved %d tweets", removed)
}

func cleanupFriends(client *twitter.Client) {
	removed := 0
	for {
		friendIds, _, _ := client.Friends.IDs(&twitter.FriendIDParams{})
		ids := friendIds.IDs
		if len(ids) == 0 {
			return
		}

		for _, friendId := range ids {
			client.Friendships.Destroy(&twitter.FriendshipDestroyParams{
				UserID: friendId,
			})
			removed += 1
		}
	}
	log.Printf("Remmoved %d friends", removed)
}

func cleanupFavorites(client *twitter.Client) {
	removed := 0
	for {
		tweets, _, _ := client.Favorites.List(&twitter.FavoriteListParams{})
		if len(tweets) == 0 {
			return
		}

		for _, tweet := range tweets {
			client.Favorites.Destroy(&twitter.FavoriteDestroyParams{
				ID: tweet.ID,
			})
			removed += 1
		}
	}
	log.Printf("Remmoved %d favorites", removed)
}
