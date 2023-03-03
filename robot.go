package robotic

import (
	"encoding/json"
	"errors"
	"github.com/glide-im/glide/pkg/auth"
	"github.com/glide-im/glide/pkg/auth/jwt_auth"
	"github.com/glide-im/glide/pkg/conn"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/timingwheel"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants/v2"
	"sync"
	"sync/atomic"
	"time"
)

type MessageInterceptor interface {
	intercept() bool
}

var timer = timingwheel.NewTimingWheel(time.Millisecond*500, 3, 20)

var dialer = websocket.Dialer{
	HandshakeTimeout:  time.Second * 3,
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

type RobotOptions struct {
}

type Robot struct {
	co conn.Connection

	msg chan *messages.GlideMessage

	seq       int64
	logCh     chan struct{}
	heartbeat *timingwheel.Task
	wg        *sync.WaitGroup

	Rec chan *messages.GlideMessage
}

func NewRobot(wsUrl string) (*Robot, error) {

	c, _, err := dialer.Dial(wsUrl, nil)
	if err != nil {
		return nil, err
	}
	connection := conn.NewWsConnection(c, &conn.WsServerOptions{
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	})

	return &Robot{
		seq:   1,
		logCh: make(chan struct{}),
		Rec:   make(chan *messages.GlideMessage, 100),
		wg:    &sync.WaitGroup{},
		msg:   make(chan *messages.GlideMessage, 100),
		co:    connection,
	}, nil
}

func (r *Robot) receive() {

	for {

		bytes, err := r.co.Read()
		if err != nil {
			logger.ErrE("receive message error", err)
			break
		}

		m := messages.NewEmptyMessage()
		err = messages.JsonCodec.Decode(bytes, m)
		if err != nil {
			logger.E("decode message error", err)
			continue
		}

		logger.I("received: %s", m)

		select {
		case r.Rec <- m:
		default:
			logger.W("too message to handle")
		}

		if err != nil {
			logger.ErrE("ants pool error", err)
		}
	}
	_ = r.Close()
}

func (r *Robot) send() {

	r.heartbeat = timer.After(time.Second * 28)

	for {

		select {
		case <-r.heartbeat.C:
			m := messages.NewMessage(0, messages.ActionHeartbeat, nil)
			if r.write(m) != nil {
				goto END
			}
		case m := <-r.msg:
			if r.write(m) != nil {
				goto END
			}
		}
	}
END:
}

func (r *Robot) write(message *messages.GlideMessage) error {

	r.heartbeat.Cancel()
	r.heartbeat = timer.After(time.Second * 28)

	m, _ := json.Marshal(message)
	logger.I("write msg: %s", string(m))

	encode, err := messages.JsonCodec.Encode(message)
	if err != nil {
		logger.ErrE("encode msg error", err)
		return nil
	}
	err = r.co.Write(encode)
	if err != nil {
		logger.ErrE("write msg error", err)
	}
	return err
}

func (r *Robot) Enqueue(m *messages.GlideMessage) error {

	if m.Seq == 0 {
		m.SetSeq(r.nextSeq())
	}
	select {
	case r.msg <- m:
	default:
		return errors.New("too many messages to enqueue, the msg queue is full")
	}
	return nil
}

func (r *Robot) BlockSend(m *messages.GlideMessage) (error, int64) {
	r.msg <- m
	return nil, 1
}

func (r *Robot) Run() error {

	go func() {
		defer func() {
			r.wg.Done()
			err := recover()
			if err != nil {
				logger.ErrE("send panic", err.(error))
			}
		}()

		r.send()
	}()

	go func() {

		defer func() {
			r.wg.Done()
			err := recover()
			if err != nil {
				logger.ErrE("receive panic", err.(error))
			}
		}()

		r.receive()
	}()

	return nil
}

func (r *Robot) Close() error {
	return r.co.Close()
}

func (r *Robot) nextSeq() int64 {
	return atomic.AddInt64(&r.seq, 1)
}

type BotX struct {
	bot   *Robot
	token string
	Id    string
	h     func(m *messages.GlideMessage)
	cmH   func(m *messages.GlideMessage, cm *messages.ChatMessage)
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

func (b BotX) Send(to string, action messages.Action, data interface{}) error {
	m := messages.NewMessage(0, action, data)
	m.To = to
	return b.bot.Enqueue(m)
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
			if authResult.Action == ActionApiFailed {
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

		if b.cmH != nil {
			b.cmH(m, &chatMsg)
		}
	}
	if b.h != nil {
		b.h(m)
	}
}
