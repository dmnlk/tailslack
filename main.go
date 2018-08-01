package main

import (
	"github.com/nlopes/slack"
	"os"
	"log"
			"github.com/fatih/color"
	"fmt"
	"time"
	"github.com/patrickmn/go-cache"
)

func main() {
	token := os.Getenv("SLACK_API_TOKEN")
	api := slack.New(token)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	api.SetDebug(false)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	cache := cache.New(cache.NoExpiration, cache.NoExpiration)

	for msg := range rtm.IncomingEvents {
		//fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {


		case *slack.MessageEvent:
		//	pp.Println(ev)
		// subtype == bot_message
			if len(ev.Text) == 0 {
				continue
			}
			if len(ev.User) == 0 {
				continue
			}
			user, err := getUserInfo(cache, api, ev.User)
			if err != nil {
				continue
			}
			color.Set(color.FgGreen)
			fmt.Print(time.Now().Format("2006/01/02 Mon 15:04:05") + ":")

			channel, err := getChannelInfo(cache, api, ev.Channel)
			if err != nil {
				continue
			}
			color.Set(color.FgHiYellow)
			fmt.Print(channel.Name + ":")
			color.Set(color.FgHiMagenta)
			fmt.Print(user.Name + ":" )
			color.White(ev.Text)
		case *slack.InvalidAuthEvent:
			os.Exit(0)
			return

		default:

			// Ignore other events..
			// fmt.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}

func getUserInfo(c* cache.Cache, api* slack.Client, user_id string) (*slack.User, error)  {
	key :=  "user-" + user_id
	if v, f := c.Get(key);  f {
		return v.(*slack.User), nil
	}
	user, err :=api.GetUserInfo(user_id)
	if err != nil {
		return nil , err
	}
	c.Set(key, user, cache.NoExpiration)
	return  user, nil
}

func getChannelInfo(c* cache.Cache, api* slack.Client, channel_id string) (*slack.Channel, error)  {
	key :=  "channel-" + channel_id
	if v, f := c.Get(key);  f {
		return v.(*slack.Channel), nil
	}
	channel, err :=api.GetChannelInfo(channel_id)
	if err != nil {
		return nil , err
	}
	c.Set(key, channel, cache.NoExpiration)
	return  channel, nil
}