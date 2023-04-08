package robotic

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTeg(t *testing.T) {

	cmd, err := NewCommand("-", "hello", func(message *messages.ChatMessage, value string) error {
		t.Log("value", value)
		return nil
	})
	assert.Nil(t, err)
	t.Log(cmd.handle(&messages.ChatMessage{Content: "#hello 123"}))
}
