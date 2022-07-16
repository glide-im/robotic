package robotic

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"sync"
	"time"
)

type Options struct {
	Ticket string
}

// MessageHandler used to handle the message that robot received.
type MessageHandler func(g gate.Gateway, m *messages.GlideMessage)

// RobotConnection .
type RobotConnection struct {
	info    *gate.Info
	gateway gate.Gateway

	exitSignal chan struct{}
	exitOnce   sync.Once
	messageCh  chan *messages.GlideMessage

	handler MessageHandler
}

func NewRobotConnection(g gate.Gateway, handler MessageHandler, opts *Options) (*RobotConnection, error) {
	if g == nil {
		return nil, errors.New("the gateway is nil")
	}
	if handler == nil {
		return nil, errors.New("the handler is nil")
	}
	return &RobotConnection{
		info: &gate.Info{
			AliveAt:      time.Now().Unix(),
			ConnectionAt: time.Now().Unix(),
			Gateway:      "",
		},
		handler:   handler,
		gateway:   g,
		messageCh: make(chan *messages.GlideMessage, 100),
	}, nil
}

func (r *RobotConnection) SetID(id gate.ID) {
	r.info.ID = id
}

func (r *RobotConnection) IsRunning() bool {
	return true
}

func (r *RobotConnection) EnqueueMessage(m *messages.GlideMessage) error {

	select {
	case r.messageCh <- m:
	default:
		return errors.New("too many messages, the robot is overload")
	}
	return nil
}

func (r *RobotConnection) Exit() {
	if r.info.ID != "" && r.gateway != nil {
		_ = r.gateway.ExitClient(r.info.ID)
	}
	r.SetID("")
	r.gateway = nil
}

func (r *RobotConnection) Run() {

	r.exitSignal = make(chan struct{})
	r.exitOnce = sync.Once{}
	r.exitOnce.Do(func() {
		close(r.exitSignal)
	})

	go func() {
		defer func() {
			err, ok := recover().(error)
			if ok {
				logger.ErrE("handle msg panic", err)
			}
		}()

		for {
			select {
			case <-r.exitSignal:
				goto END
			case m := <-r.messageCh:
				logger.D("handle msg: %s", m)
				if m == nil {
					goto END
				}
				r.handler(r.gateway, m)
			}
		}
	END:
		logger.D("robot %s exit", r.info.ID)
	}()
	logger.D("robot %s running", r.info.ID)
}

func (r *RobotConnection) GetInfo() gate.Info {
	return *r.info
}
