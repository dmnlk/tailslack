package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/nlopes/slack"
	"github.com/patrickmn/go-cache"
	"gopkg.in/kyokomi/emoji.v1"
	"log"
	"os"
	"strings"
	"time"
	"github.com/deckarep/gosx-notifier"
)

func main() {
	token := os.Getenv("SLACK_API_TOKEN")
	api := slack.New(token)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	api.SetDebug(false)

	user, err := api.AuthTest()
	if err != nil {
		color.Set(color.FgRed)
		log.Fatal("OAuthError!")
	}
	ownUserId := user.UserID
	fmt.Printf("your user id is %s\n", ownUserId)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	cache := cache.New(cache.NoExpiration, cache.NoExpiration)
	scaner := bufio.NewScanner(os.Stdin)
	go func() {
		for scaner.Scan() {
			input := strings.Fields(scaner.Text())

			if len(input) == 0 {
				// like tailf
				continue
			}
			if len(input) != 3 {
				fmt.Println("argument count error")
				continue
			}
			if input[0] != "/post" {
				fmt.Println("invalid command")
				continue
			}

			//postMessage
			params := slack.PostMessageParameters{}
			params.AsUser = true
			channelID, timestamp, err := api.PostMessage(input[1], input[2], params)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
			fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
		}
	}()
	for msg := range rtm.IncomingEvents {
		//fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:
			//pp.Println(ev)
			// subtype == bot_message
			var user_name = ""
			text := ev.Text
			if len(text) == 0 {
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
				if len(user_name) == 0 {
					user_name = user.Name
				}
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
			var orgText  = text
			formatedText := emoji.Sprint(text)
			color.White(formatedText)
			if strings.Contains(text, ownUserId) || strings.Contains(text, "<!here>") || strings.Contains(text, "<!channel>") {
				notification := gosxnotifier.NewNotification(orgText) // don't show message why?
				notification.Title = channelName
				notification.Subtitle = user_name
				notification.AppIcon = "icon.png"
				notification.Sound = gosxnotifier.Default
				notification.Push()
			}


		case *slack.InvalidAuthEvent:
			color.Set(color.FgRed)
			fmt.Println("Auth Error!! Please Check SLACK_API_TOKEN")
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
