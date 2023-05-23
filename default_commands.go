package robotic

import "fmt"

func CommandPing(x *BotX) *Command {
	cmd, _ := NewCommand2(0, "ping", "test robot available", func(cm *ResolvedChatMessage, value string) error {
		return x.Reply(cm, MessageTypeText, "pong")
	})
	return cmd
}

func CommandHelp(x *BotX) *Command {
	cmd, _ := NewCommand2(0, "help", "show command list", func(cm *ResolvedChatMessage, value string) error {
		commandList := "Commands:\n"
		for _, c := range x.Commands {
			commandList += fmt.Sprintf("- %s: %s\n", c.Name, c.Desc)
		}
		return x.Reply(cm, MessageTypeMarkdown, commandList)
	})
	return cmd
}
