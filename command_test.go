package robotic

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTeg(t *testing.T) {

	cmd, err := NewCommand(Role(0), "hello", func(message *ResolvedChatMessage, value string) error {
		t.Log("value", value)
		return nil
	})
	assert.Nil(t, err)
	t.Log(cmd.handle(&ResolvedChatMessage{
		Origin: &messages.GlideMessage{},
		ChatMessage: &messages.ChatMessage{
			Content: "#hello world 123 ",
		},
	}))
}
