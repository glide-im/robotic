package robotic

import (
	"errors"
	"github.com/glide-im/glide/pkg/messages"
	"io/ioutil"
	"net/http"
)

type Response struct {
	Msg  string
	Code int
	Data struct {
		Mid int64
	}
}

func GetMid(token string) (error, int64) {
	request, err := http.NewRequest("POST", "http://api.glide-im.pro/api/msg/id", nil)
	if err != nil {
		return err, 0
	}
	request.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err, 0
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status), 0
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, 0
	}
	r := Response{}
	err = messages.JsonCodec.Decode(bytes, &r)
	if err != nil {
		return err, 0
	}

	if r.Code != 100 {
		return errors.New(r.Msg), 0
	}
	return nil, r.Data.Mid
}
