package robotic

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/glide-im/glide/pkg/auth"
	"github.com/glide-im/glide/pkg/auth/jwt_auth"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/google/uuid"
	"github.com/panjf2000/ants/v2"
	"strings"
	"time"
)

type BotX struct {
	bot   *Robot
	token string
	Id    string
	h     func(m *messages.GlideMessage)
	cmH   func(m *messages.GlideMessage, cm *messages.ChatMessage)

	Commands []*Command
}

type ResolvedChatMessage struct {
	Origin      *messages.GlideMessage
	ChatMessage *messages.ChatMessage
}

func NewBotX(wsUrl, token string) *BotX {

	robot, err := NewRobot(wsUrl)
	if err != nil {
		panic(err)
	}

	x := &BotX{
		bot:   robot,
		token: token,
	}
	return x
}

func (b *BotX) Login() {

}

func (b *BotX) Send(to string, action messages.Action, data interface{}) error {
	m := messages.NewMessage(0, action, data)
	m.To = to
	return b.bot.Enqueue(m)
}

func (b *BotX) AddCommand(command *Command) error {
	if command.regex == nil {
		err := command.compileRe()
		if err != nil {
			return err
		}
	}
	b.Commands = append(b.Commands, command)
	return nil
}

func (b *BotX) Start(h func(m *messages.GlideMessage)) error {
	b.h = h

	pool, err := ants.NewPool(1_0000,
		ants.WithNonblocking(true),
		ants.WithPreAlloc(false),
		ants.WithNonblocking(true),
		ants.WithPanicHandler(func(i interface{}) {
			logger.ErrE("handel message panic", i.(error))
		}))

	if err != nil {
		return err
	}

	authMsg := messages.NewMessage(b.bot.nextSeq(), ActionApiAuth, &auth.Token{Token: b.token})
	err = b.bot.Enqueue(authMsg)
	if err != nil {
		return err
	}

	authResultCh := make(chan *messages.GlideMessage)
	authSeq := authMsg.GetSeq()

	errCh := make(chan error, 3)

	go func() {
		defer func() {
			err, ok := recover().(error)
			if ok {
				errCh <- err
			}
		}()

		for m := range b.bot.Rec {
			e := pool.Submit(func() {
				if m.GetSeq() == authSeq && authSeq > 0 {
					authResultCh <- m
					return
				}
				b.onReceive(m)
			})
			if e != nil {
				logger.ErrE("ants pool submit error", e)
			}
		}
	}()

	go func() {
		select {
		case authResult := <-authResultCh:
			if authResult.Action == string(ActionApiFailed) {
				errCh <- errors.New(authResult.Data.String())
				break
			}
			info := jwt_auth.Response{}
			err = authResult.Data.Deserialize(&info)
			if err != nil {
				errCh <- err
			}
			b.token = info.Token
			b.Id = info.Uid
			logger.D("login success: %s", info.Uid)
		case <-timer.After(time.Second * 10).C:
			errCh <- errors.New("login timeout")
		}
		authSeq = -1
		close(authResultCh)
	}()

	_ = b.bot.Run()

	err = <-errCh

	return err
}

func (b *BotX) HandleChatMessage(h func(m *messages.GlideMessage, cm *messages.ChatMessage)) {
	b.cmH = h
}

func (b *BotX) onReceive(m *messages.GlideMessage) {
	switch m.GetAction() {
	case ActionNotifyKickOut:
		_ = b.bot.Close()
		return
	case ActionChatMessageResend:
		fallthrough
	case ActionChatMessage, ActionGroupMessage:
		chatMsg := messages.ChatMessage{}
		err := m.Data.Deserialize(&chatMsg)
		if err != nil {
			logger.ErrE("decode chat msg error", err)
			return
		}

		ack := &messages.AckRequest{
			Seq:  m.GetSeq(),
			Mid:  chatMsg.Mid,
			From: b.Id,
		}
		_ = b.Send(m.From, ActionAckRequest, ack)

		for _, command := range b.Commands {
			if command.handle(&ResolvedChatMessage{
				Origin:      m,
				ChatMessage: &chatMsg,
			}) {
				return
			}
		}

		if b.cmH != nil {
			b.cmH(m, &chatMsg)
		}
	}
	if b.h != nil {
		b.h(m)
	}
}

func (b *BotX) Reply(originMessage *ResolvedChatMessage, messageType int32, content interface{}) error {

	from := b.Id
	to := originMessage.ChatMessage.From
	action := ActionChatMessage

	contentFormat := "%s"

	if originMessage.Origin.Action == string(ActionGroupMessage) {
		action = ActionGroupMessage
		to = originMessage.ChatMessage.To
		contentFormat = "@" + originMessage.ChatMessage.From + " %s"
	} else {
		action = ActionChatMessage
		to = originMessage.ChatMessage.From
	}

	uid, i := newUid()

	cnt, ok := content.(string)
	if !ok {
		bs, err := json.Marshal(content)
		if err != nil {
			return err
		}
		cnt = string(bs)
	}

	chatMessage := &messages.ChatMessage{
		CliMid:  uid,
		Mid:     i,
		From:    from,
		To:      to,
		Type:    messageType,
		Content: fmt.Sprintf(contentFormat, cnt),
		SendAt:  time.Now().UnixMilli(),
	}

	return b.Send(to, action, chatMessage)
}

func newUid() (string, int64) {

	id2, _ := uuid.NewUUID()
	idstr2 := strings.ReplaceAll(strings.ToUpper(id2.String()), "-", "")

	return idstr2, int64(id2.ID())
}
