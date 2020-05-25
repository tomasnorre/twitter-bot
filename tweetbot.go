package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/adam-lavrik/go-imath/ix"
	"github.com/dghubble/oauth1"
	"github.com/tomasnorre/go-twitter/twitter"
	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Twitter TwitterConf `yaml:"twitter"`
}

type TwitterConf struct {
	OauthAccessToken       string   `yaml:"oauth_access_token"`
	OauthAccessTokenSecret string   `yaml:"oauth_access_token_secret"`
	ConsumerKey            string   `yaml:"consumer_key"`
	ConsumerSecret         string   `yaml:"consumer_secret"`
	Hash                   []string `yaml:"hash"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// WriteToFile will print any string of text to a file safely by
// checking for errors and syncing at the end.
func WriteToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	check(err)
	return file.Sync()
}

func main() {

	file, err := os.Open("settings.yaml")
	if err != nil {
		panic(err)
	}

	filecontent, err := ioutil.ReadAll(file)
	check(err)

	var conf Configuration
	err = yaml.Unmarshal(filecontent, &conf)
	check(err)

	fileReference, err := os.Open("lastTweetId")
	check(err)

	lastTweetId, err := ioutil.ReadAll(fileReference)
	check(err)

	config := oauth1.NewConfig(conf.Twitter.ConsumerKey, conf.Twitter.ConsumerSecret)
	token := oauth1.NewToken(conf.Twitter.OauthAccessToken, conf.Twitter.OauthAccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)
	lastTweetIdValue, err := strconv.ParseInt(string(lastTweetId), 0, 64)
	check(err)

	out := make(chan *twitter.Search)

	for _, h := range conf.Twitter.Hash {
		go func(hash string) {
			// Search Tweets
			search, _, err := client.Search.Tweets(&twitter.SearchTweetParams{
				Query:      hash,
				Count:      5,
				ResultType: "recent",
				SinceID:    lastTweetIdValue,
			})
			check(err)
			out <- search
		}(h)
	}

	var tweetIds []int

	for i := 0; i < len(conf.Twitter.Hash); i++ {
		search := <-out

		for _, tweet := range search.Statuses {
			if tweet.RetweetedStatus != nil {
				continue
			}
			tweetIds = append(tweetIds, int(tweet.ID))
			fmt.Println(tweet.ID, tweet.Text)

			var statusRetweetParam *twitter.StatusRetweetParams
			client.Statuses.Retweet(tweet.ID, statusRetweetParam)

		}
	}
	close(out)

	// Write latestTweetId to know where to start on next execution.
	if 0 < len(tweetIds) {
		latestTweetId := strconv.Itoa(ix.MaxSlice(tweetIds))
		WriteToFile("lastTweetId", latestTweetId)
	}
}
