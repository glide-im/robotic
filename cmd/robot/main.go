package main

import (
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/robotic"
	"github.com/glide-im/robotic/config"
)

func main() {

	conf, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	botX := robotic.NewBotX(conf.Bot.Ws, conf.Bot.Token)
	botX.HandleChatMessage(func(m *messages.GlideMessage, cm *messages.ChatMessage) {
		logger.I("handler chat message >> %s", m.GetAction())
	})

	_ = botX.AddCommand(robotic.CommandPing(botX))
	_ = botX.AddCommand(robotic.CommandHelp(botX))

	err = botX.Start(nil)
	panic(err)
}
