package robotic

import "github.com/glide-im/glide/pkg/messages"

const (
	ActionHello               messages.Action = "hello"
	ActionHeartbeat           messages.Action = "heartbeat"
	ActionNotifyUnknownAction messages.Action = "notify.unknown.action"

	ActionChatMessage       messages.Action = "message.chat"
	ActionChatMessageResend messages.Action = "message.chat.resend"
	ActionGroupMessage      messages.Action = "message.group"
	ActionMessageFailed     messages.Action = "message.failed.send"

	ActionNotifyNeedAuth      messages.Action = "notify.auth"
	ActionNotifyKickOut       messages.Action = "notify.kickout"
	ActionNotifyNewContact    messages.Action = "notify.contact"
	ActionNotifyGroup         messages.Action = "notify.group"
	ActionNotifyAccountLogin  messages.Action = "notify.login"
	ActionNotifyAccountLogout messages.Action = "notify.logout"
	ActionNotifyError         messages.Action = "notify.error"

	ActionAckRequest  messages.Action = "ack.request"
	ActionAckGroupMsg messages.Action = "ack.group.msg"
	ActionAckMessage  messages.Action = "ack.message"
	ActionAckNotify   messages.Action = "ack.notify"

	ActionApiAuth    messages.Action = "api.auth"
	ActionApiFailed  messages.Action = "api.failed"
	ActionApiSuccess messages.Action = "api.success"

	ActionClientCustom messages.Action = "message.cli"

	NotifyKickOut messages.Action = "notify.kickout"
	AckOffline    messages.Action = "ack.offline"
)
