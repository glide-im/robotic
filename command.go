package robotic

import (
	"errors"
	"fmt"
	"github.com/glide-im/glide/pkg/messages"
	"regexp"
	"strconv"
)

const (
	CommandPrefix    = "#"
	CommandMaxLength = 10

	commandRegexPTemp = "^%s([a-zA-Z]{1,%d})$|^%s([a-zA-Z]{1,%d}) ([^ ]+)$"
)

var commandNameRegexp = regexp.MustCompile(fmt.Sprintf("^[a-zA-Z]{1,%d}$", CommandMaxLength))

type CommandHandler func(message *messages.ChatMessage, value string) error

type Command struct {
	Role   Role
	Name   string
	Handle CommandHandler

	regex *regexp.Regexp
}

func NewCommand(role Role, name string, handle CommandHandler) (*Command, error) {

	c := &Command{
		Role:   role,
		Name:   name,
		Handle: handle,
		regex:  nil,
	}
	err := c.compileRe()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Command) compileRe() error {
	if !commandNameRegexp.MatchString(c.Name) {
		return errors.New("command name must be: A-Z, a-z, 1<len<" + strconv.Itoa(CommandMaxLength))
	}

	re, err := regexp.Compile(fmt.Sprintf(commandRegexPTemp, CommandPrefix, CommandMaxLength, CommandPrefix, CommandMaxLength))
	if c.regex == nil {
		c.regex = re
	}
	return err
}

func (c Command) handle(message *messages.ChatMessage) bool {
	match := c.regex.FindStringSubmatch(message.Content)
	if len(match) != 0 {
		if c.Name == match[2] || c.Name == match[1] {
			_ = c.Handle(message, match[3])
			return true
		}
	}
	return false
}
