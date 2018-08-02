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
				user_name = user.Profile.DisplayName
			}

			color.Set(color.FgGreen)
			fmt.Print(time.Now().Format("2006/01/02 Mon 15:04:05") + ":")

			var channelName = ""
			channel, err := getChannelInfo(cache, api, ev.Channel)
			if err != nil {
				color.Set(color.FgRed)
				channelName = "[DM] by " + user_name
			} else {
				color.Set(color.FgHiYellow)
				channelName = channel.Name
			}

			fmt.Print(channelName + ":")
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

func getUserInfo(c *cache.Cache, api *slack.Client, userId string) (*slack.User, error) {
	key := "user-" + userId
	if v, f := c.Get(key); f {
		return v.(*slack.User), nil
	}
	user, err := api.GetUserInfo(userId)
	if err != nil {
		return nil, err
	}
	c.Set(key, user, cache.NoExpiration)
	return user, nil
}

func getChannelInfo(c *cache.Cache, api *slack.Client, channelId string) (*slack.Channel, error) {
	key := "channel-" + channelId
	if v, f := c.Get(key); f {
		return v.(*slack.Channel), nil
	}
	channel, err := api.GetChannelInfo(channelId)
	if err != nil {
		return nil, err
	}
	c.Set(key, channel, cache.NoExpiration)
	return channel, nil
}
