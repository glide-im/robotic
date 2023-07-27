package robotic

import "testing"

func init() {
	ApiBaseUrl = "https://intercom.ink"
}

func TestLogin(t *testing.T) {
	a, err := Login("", "")
	if err != nil {
		panic(err)
	}
	t.Log(a.Token, a.NickName, a.Credential)
}

func TestGuestLogin(t *testing.T) {
	a, err := GuestLogin()
	if err != nil {
		panic(err)
	}
	t.Log(a.Token, a.NickName, a.Credential)
}
