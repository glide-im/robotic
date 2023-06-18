package robotic

import "testing"

func TestLogin(t *testing.T) {
	a, err := Login("", "")
	if err != nil {
		panic(err)
	}
	t.Log(a.Token, a.NickName, a.Credential)
}
