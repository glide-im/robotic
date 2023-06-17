package robotic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/robotic/config"
	"io"
	"io/ioutil"
	"net/http"
)

var token string

type Response struct {
	Msg  string
	Code int
	Data *Data
}

type Data struct {
	bytes []byte
}

func (d *Data) UnmarshalJSON(i []byte) error {
	d.bytes = i
	return nil
}

type Credential struct {
	Version    int64  `json:"version,omitempty"`
	Credential string `json:"credential,omitempty"`
}
type AuthResponse struct {
	Token      string      `json:"token"`
	Servers    []string    `json:"servers"`
	NickName   string      `json:"nick_name"`
	Uid        int64       `json:"uid"`
	Status     int         `json:"status"`
	Credential *Credential `json:"credential"`
}

func RequestApi(method string, url string, body interface{}) (*Response, error) {
	var b io.Reader
	if body != nil {
		marshal, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		b = bytes.NewBuffer(marshal)
	}
	request, err := http.NewRequest(method, url, b)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("%d %s", resp.StatusCode, resp.Status))
	}
	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := Response{}
	err = messages.JsonCodec.Decode(rb, &r)
	if err != nil {
		return nil, err
	}

	if r.Code != 100 {
		return nil, errors.New(fmt.Sprintf("%d %s", r.Code, r.Msg))
	}
	return &r, nil
}

func Login(account, password string) (*AuthResponse, error) {
	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/auth/signin_v2", c.Api.BaseUrl)
	resp, err := RequestApi("POST", url, struct {
		Device   int
		Password string
		Email    string
	}{
		0, password, account,
	})
	if err != nil {
		return nil, err
	}
	auth := AuthResponse{}
	err = json.Unmarshal(resp.Data.bytes, &auth)
	if err != nil {
		return nil, err
	}
	return &auth, nil
}
