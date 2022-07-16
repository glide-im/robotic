package robotic

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
)

type ProxyRobot struct {
}

func (p *ProxyRobot) SetID(id gate.ID) {

}

func (p *ProxyRobot) IsRunning() bool {
	return true
}

func (p *ProxyRobot) EnqueueMessage(message *messages.GlideMessage) error {

	return nil
}

func (p *ProxyRobot) Exit() {

}

func (p *ProxyRobot) Run() {

}

func (p *ProxyRobot) GetInfo() gate.Info {
	return gate.Info{}
}
