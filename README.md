## Robotic

### 安装

```shell
go get github.com/glide-im/robotic
```

### 使用

这里实现了一个简单的 echo 机器人

```go
botX := robotic.NewBotX("ws://bot_server.address", "Token")
botX.HandleChatMessage(func(m *messages.GlideMessage, cm *messages.ChatMessage) {
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
            _ := botX.Send(cm.From, robotic.ActionChatMessage, &replyMsg)
        }()
    }
})
err = botX.Start(nil)
panic(err)
```