package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/nlopes/slack"
	"github.com/patrickmn/go-cache"
	"log"
	"os"
	"time"
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
			//pp.Println(ev)
			// subtype == bot_message
			var user_name = ""
			if len(ev.Text) == 0 {
				continue
			}
			if len(ev.User) == 0 {
				if len(ev.Username) == 0 {
					continue
				}
				// bot name
				user_name = ev.Username
			}
			if len(user_name) == 0 {
				user, err := getUserInfo(cache, api, ev.User)

				if err != nil {
					continue
				}
				user_name = user.Name
			}

			color.Set(color.FgGreen)
			fmt.Print(time.Now().Format("2006/01/02 Mon 15:04:05") + ":")

			var chan_name = ""
			channel, err := getChannelInfo(cache, api, ev.Channel)
			if err != nil {
				color.Set(color.FgRed)
				chan_name = "[DM] by " + user_name
			} else {
				color.Set(color.FgHiYellow)
				chan_name = channel.Name
			}

			fmt.Print(chan_name + ":")
			color.Set(color.FgHiMagenta)
			fmt.Print(user_name + ":")
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

func getUserInfo(c *cache.Cache, api *slack.Client, user_id string) (*slack.User, error) {
	key := "user-" + user_id
	if v, f := c.Get(key); f {
		return v.(*slack.User), nil
	}
	user, err := api.GetUserInfo(user_id)
	if err != nil {
		return nil, err
	}
	c.Set(key, user, cache.NoExpiration)
	return user, nil
}

func getChannelInfo(c *cache.Cache, api *slack.Client, channel_id string) (*slack.Channel, error) {
	key := "channel-" + channel_id
	if v, f := c.Get(key); f {
		return v.(*slack.Channel), nil
	}
	channel, err := api.GetChannelInfo(channel_id)
	if err != nil {
		return nil, err
	}
	c.Set(key, channel, cache.NoExpiration)
	return channel, nil
}
