package robotic

import (
	"errors"
	"github.com/glide-im/glide/pkg/messages"
)

type Perm int64

type Role int64

func (r Role) isAllow(perms ...Perm) bool {
	for _, m := range perms {
		if r.isDenied(m) {
			return false
		}
	}
	return true
}

func (r Role) isDenied(perm Perm) bool {
	b := perm >> r
	return b&1 != 1
}

func (r Role) IsApply(other Role) bool {
	result := r & other
	return result != 0
}

func (r Role) allow(perm Perm) {
	// todo
}

type UserId interface{}

var RoleController = &roleController{
	defaultRole: 1,
	userRole:    map[UserId]Role{},
}

type RoleControllerInterface interface {
	GetRules(id UserId) Role

	Apply(id UserId, role Role) error

	SetUserRole(id UserId, name Perm, enable bool)
}

type roleController struct {
	defaultRole Role
	userRole    map[UserId]Role
}

func (r *roleController) GetRules(id UserId) Role {

	roles, ok := r.userRole[id]
	if ok {
		return roles
	}
	return r.defaultRole
}

func (r *roleController) SetUserRole(id UserId, name Perm, enable bool) {
	role := r.userRole[id]
	if enable {
		role.isAllow(name)
	} else {
		role.isDenied(name)
	}
}

func (r *roleController) Apply(id UserId, role Role) error {
	rules := r.GetRules(id)
	if rules.IsApply(role) {
		return nil
	}
	return errors.New("permission denied")
}

func GetUserRoleFromMessage(message *messages.ChatMessage) Role {
	return RoleController.GetRules(message.From)
}
