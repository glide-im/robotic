package main

import (
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/robotic"
	"github.com/glide-im/robotic/config"
	"time"
)

func main() {

	conf, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	botX := robotic.NewBotX(conf.Bot.Ws, conf.Bot.Token)
	botX.HandleChatMessage(func(m *messages.GlideMessage, cm *messages.ChatMessage) {
		if m.GetAction() == messages.ActionChatMessage {
			err, mid := robotic.GetMid(conf.Bot.Token)
			if err != nil {
				logger.ErrE("get mid error", err)
				return
			}
			echo := messages.ChatMessage{
				Mid:     mid,
				From:    cm.To,
				To:      cm.From,
				Type:    cm.Type,
				Content: cm.Content,
				SendAt:  time.Now().Unix(),
			}
			_ = botX.Send(m.From, messages.ActionChatMessage, &echo)
		}
	})

	err = botX.Start(nil)
	panic(err)
}
