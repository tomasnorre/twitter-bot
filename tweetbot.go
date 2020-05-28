package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Printf("could not close file %s: %s", file.Name(), err)
	}
}

// WriteToFile will print any string of text to a file safely by
// checking for errors and syncing at the end.
func WriteToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not open file: %s", err)
	}
	defer closeFile(file)

	_, err = io.WriteString(file, data)
	if err != nil {
		return fmt.Errorf("could not write string: %s", err)
	}
	return nil
}

func getLastTweetID() (int64, error) {
	lastTweetIdFile, err := os.Open("lastTweetId")
	if err != nil {
		return 0, fmt.Errorf("could not open lastTweetId: %s", err)
	}
	defer closeFile(lastTweetIdFile)

	lastTweetId, err := ioutil.ReadAll(lastTweetIdFile)
	if err != nil {
		return 0, fmt.Errorf("could not read lastTweetId content: %s", err)
	}

	lastTweetIdValue, err := strconv.ParseInt(string(lastTweetId), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse lastTweetId content as int64: %s", err)
	}

	return lastTweetIdValue, nil
}

func main() {

	file, err := os.Open("settings.yaml")
	if err != nil {
		log.Fatalf("could not open file: %s", err)
	}
	defer closeFile(file)

	filecontent, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("could not read settings.yaml content: %s", err)
	}

	var conf Configuration
	err = yaml.Unmarshal(filecontent, &conf)
	if err != nil {
		log.Fatalf("could not unmarshal settings.yaml content: %s", err)
	}

	var lastTweetID int64
	if id, err := getLastTweetID(); err != nil {
		log.Printf("could not get last tweet id: %s", err)
	} else {
		lastTweetID = id
	}

	config := oauth1.NewConfig(conf.Twitter.ConsumerKey, conf.Twitter.ConsumerSecret)
	token := oauth1.NewToken(conf.Twitter.OauthAccessToken, conf.Twitter.OauthAccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	out := make(chan *twitter.Search)

	for _, h := range conf.Twitter.Hash {
		go func(hash string) {
			// Search Tweets

			params := &twitter.SearchTweetParams{
				Query:      hash,
				Count:      5,
				ResultType: "recent",
			}

			if lastTweetID != 0 {
				params.SinceID = lastTweetID
			}

			search, _, err := client.Search.Tweets(params)
			if err != nil {
				log.Printf("could not search tweets for hash '%s': %s", hash, err)
				// a search did not execute properly, send an empty object so we don't deadlock
				out <- &twitter.Search{}
				return
			}
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
			_, _, err = client.Statuses.Retweet(tweet.ID, statusRetweetParam)
			if err != nil {
				log.Printf("could not retweet for hash '%s': %s", conf.Twitter.Hash[i], err)
				continue
			}
		}
	}
	close(out)

	// Write latestTweetId to know where to start on next execution.
	if 0 < len(tweetIds) {
		latestTweetId := strconv.Itoa(ix.MaxSlice(tweetIds))
		err := WriteToFile("lastTweetId", latestTweetId)
		if err != nil {
			log.Fatalf("could not write lastTweetId to file: %s", err)
		}
	}
}
