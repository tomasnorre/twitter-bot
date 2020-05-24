# Twitter Bot

This is a small and very simple twitter bot, which retweets based on search queries.

## How to run this bot

Add a `settings.yaml` file, same structure as `settings.example.yaml` and add your API Tokens etc. plus the search strings you want to retweet.

When done, you can run it by doing:

```shell
$ go run tweetbot.go
```

You can also compile a binary, depending on where it has to run

```shell 
$ go build tweetbot.go
```

or if you want it to run on you raspberry pi.

```shell 
env GOOS=linux GOARCH=arm GOARM=5 go build tweetbot.go
```

## Credits

Thanks to [Miloskrca](https://github.com/miloskrca) for helping with the initial development and for introducing me to Golang.


