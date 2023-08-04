package server

import (
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/protobuf/proto"
	stdhttp "net/http"
	"time"
)

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTPError code: %d message: %s", e.Code, e.Message)
}

// FromError try to convert an error to *HTTPError.
func FromError(err error) *HTTPError {
	if err == nil {
		return nil
	}
	if se := new(HTTPError); errors.As(err, &se) {
		return se
	}
	return &HTTPError{Code: 500, Message: err.Error()}
}

func errorEncoder(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
	se := FromError(err)
	codec, _ := http.CodecForRequest(r, "Accept")
	body, err := codec.Marshal(se)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/"+codec.Name())
	// todo: 判断系统错误和业务错误
	//if se.Code > 99 && se.Code < 600 {
	//	w.WriteHeader(se.Code)
	//} else {
	//	w.WriteHeader(500)
	//}
	w.WriteHeader(se.Code)
	_, _ = w.Write(body)
}

type HTTPOk struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Ts      int64       `json:"ts"`
}

func respEncoder(w stdhttp.ResponseWriter, r *stdhttp.Request, v interface{}) error {
	codec, _ := http.CodecForRequest(r, "Accept")
	messageMap := make(map[string]interface{})
	messageStr, _ := codec.Marshal(v.(proto.Message))
	_ = codec.Unmarshal(messageStr, &messageMap)

	if len(messageMap) == 1 {
		for _, vv := range messageMap {
			v = vv
		}
	}

	reply := &HTTPOk{
		Code:    200,
		Message: "success",
		Data:    v,
		Ts:      time.Now().Unix(),
	}

	data, err := codec.Marshal(reply)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/"+codec.Name())
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}
