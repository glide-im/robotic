package main

import (
	"fmt"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/robotic"
	"github.com/glide-im/robotic/config"
	"github.com/google/uuid"
	"strings"
	"time"
)

func main() {

	conf, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	botX := robotic.NewBotX(conf.Bot.Ws, conf.Bot.Token)
	botX.HandleChatMessage(func(m *messages.GlideMessage, cm *messages.ChatMessage) {
		logger.I("handler chat message >> %s", m.GetAction())
		if m.GetAction() == robotic.ActionChatMessage {
			go func() {
				replyMsg := messages.ChatMessage{
					CliMid:  uuid.New().String(),
					Mid:     0,
					From:    botX.Id,
					To:      cm.From,
					Type:    cm.Type,
					Content: cm.Content,
					SendAt:  time.Now().Unix(),
				}
				err2 := botX.Send(cm.From, robotic.ActionChatMessage, &replyMsg)
				if err2 != nil {
					logger.ErrE("send error", err2)
				}
			}()
		}
		if m.GetAction() == robotic.ActionGroupMessage {
			logger.I("Receive Group Message: %s", m.To)
			if strings.HasPrefix(cm.Content, "@"+conf.Bot.Name) {
				cnt := strings.TrimPrefix(cm.Content, "@"+conf.Bot.Name)

				go func() {

					replyMsg := messages.ChatMessage{
						CliMid:  uuid.New().String(),
						From:    botX.Id,
						To:      cm.To,
						Type:    cm.Type,
						Content: fmt.Sprintf("@%s %s", cm.From, cnt),
						SendAt:  time.Now().Unix(),
					}
					err2 := botX.Send(m.To, robotic.ActionGroupMessage, &replyMsg)
					if err2 != nil {
						logger.ErrE("send error", err2)
					}

				}()
			}
		}
	})

	err = botX.Start(nil)
	panic(err)
}
