package robotic

import (
	"encoding/json"
	"errors"
	"github.com/glide-im/glide/pkg/conn"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/timingwheel"
	"github.com/gorilla/websocket"
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

		b, _ := json.Marshal(&m)
		logger.I("received: %s", string(b))

		if m.Action == messages.ActionHeartbeat {
			_ = r.Enqueue(messages.NewMessage(0, messages.ActionHeartbeat, nil))
			return
		}

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

	r.heartbeat = timer.After(time.Second * 60)

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
	r.heartbeat = timer.After(time.Second * 60)

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
			err := recover()
			if err != nil {
				logger.ErrE("send panic", err.(error))
			}
		}()

		r.send()
	}()

	go func() {

		defer func() {
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
