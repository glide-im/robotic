package robotic

const (
	ActionHello               = "hello"
	ActionHeartbeat           = "heartbeat"
	ActionNotifyUnknownAction = "notify.unknown.action"

	ActionChatMessage       = "message.chat"
	ActionChatMessageResend = "message.chat.resend"
	ActionGroupMessage      = "message.group"
	ActionMessageFailed     = "message.failed.send"

	ActionNotifyNeedAuth      = "notify.auth"
	ActionNotifyKickOut       = "notify.kickout"
	ActionNotifyNewContact    = "notify.contact"
	ActionNotifyGroup         = "notify.group"
	ActionNotifyAccountLogin  = "notify.login"
	ActionNotifyAccountLogout = "notify.logout"
	ActionNotifyError         = "notify.error"

	ActionAckRequest  = "ack.request"
	ActionAckGroupMsg = "ack.group.msg"
	ActionAckMessage  = "ack.message"
	ActionAckNotify   = "ack.notify"

	ActionApiAuth    = "api.auth"
	ActionApiFailed  = "api.failed"
	ActionApiSuccess = "api.success"

	ActionClientCustom = "message.cli"

	NotifyKickOut = "notify.kickout"
	AckOffline    = "ack.offline"
)
