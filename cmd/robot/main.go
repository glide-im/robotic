package main

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/robotic"
	"time"
)

func main() {

	token := ""
	wsUrl := ""

	botX := robotic.NewBotX(wsUrl, token)
	botX.HandleChatMessage(func(m *messages.GlideMessage, cm *messages.ChatMessage) {
		if m.GetAction() == messages.ActionChatMessage {
			m := messages.ChatMessage{
				Mid:     0,
				From:    cm.To,
				To:      cm.From,
				Type:    cm.Type,
				Content: cm.Content,
				SendAt:  time.Now().Unix(),
			}
			_ = botX.Send(m.From, messages.ActionChatMessage, "")
		}
	})

	err := botX.Start(func(m *messages.GlideMessage) {

	})
	panic(err)
}
