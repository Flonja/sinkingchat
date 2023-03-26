package chat

import (
	"encoding/json"
	"fmt"
	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"strings"
	"time"
)

type FloatplaneChatSocket struct {
	c        *gosocketio.Client
	username string
	guid     string
	emotes   *[]Emote
	channel  string
}

func NewFloatplaneChatSocket(channel string) (*FloatplaneChatSocket, error) {
	sailsToken := Login()

	if !strings.HasPrefix(sailsToken, "s:") {
		return nil, fmt.Errorf("sailsToken doesn't start with `s:`, likely not valid")
	}

	wssUrl := "wss://chat.floatplane.com/socket.io/?__sails_io_sdk_version=0.13.8&__sails_io_sdk_platform=browser&__sails_io_sdk_language=javascript&EIO=3&transport=websocket"
	trnsprt := transport.GetDefaultWebsocketTransport()
	trnsprt.RequestHeader = http.Header{}
	trnsprt.RequestHeader.Set("Origin", "https://www.floatplane.com")
	trnsprt.RequestHeader.Set("Cookie", fmt.Sprintf("sails.sid=%v", sailsToken))
	client, err := gosocketio.Dial(wssUrl, trnsprt)
	if err != nil {
		return nil, err
	}
	socket := &FloatplaneChatSocket{c: client, channel: channel}
	d := &ResponseJoinRoom{}
	if err = ack(socket, newJoinLivestreamRequest(socket.channel), d); err != nil {
		return nil, err
	}
	if d.Success {
		socket.emotes = d.Emotes
	}
	msg := &ResponseRoomMessage{}
	if err = ack(socket, newSendLivestreamMsgRequest(socket.channel, "Initialized!"), msg); err != nil {
		return nil, err
	}
	socket.username = msg.Username
	socket.guid = msg.UserGuid

	return socket, nil
}

func (f *FloatplaneChatSocket) Username() string {
	return f.username
}

func (f *FloatplaneChatSocket) Guid() string {
	return f.guid
}

func (f *FloatplaneChatSocket) Emotes() ([]Emote, bool) {
	return *f.emotes, f.emotes != nil
}

func (f *FloatplaneChatSocket) Users() (*ResponseUserList, error) {
	d := &ResponseUserList{}
	return d, ack(f, newGetUserListRequest(f.channel), d)
}

func (f *FloatplaneChatSocket) SendMessage(message string) error {
	return f.c.Emit("post", newSendLivestreamMsgRequest(f.channel, message))
}

func (f *FloatplaneChatSocket) Close() error {
	if err := f.c.Emit("get", newLeaveLivestreamRequest(f.channel)); err != nil {
		return err
	}
	f.c.Close()
	return nil
}

func (f *FloatplaneChatSocket) Listen(function func(*ResponseRoomMessage)) error {
	return f.c.On("radioChatter", func(c *gosocketio.Channel, args interface{}) {
		d := &ResponseRoomMessage{}
		if err := mapstructure.Decode(args, d); err == nil {
			function(d)
		}
	})
}

func ack[T any](f *FloatplaneChatSocket, request *Request, out *T) error {
	resp, err := f.c.Ack(request.Method, request, time.Second*10)
	if err != nil {
		return err
	}

	r := &Response{}
	if err = json.Unmarshal([]byte(resp), r); err != nil {
		return err
	}
	if err = mapstructure.Decode(r.Body, out); err != nil {
		return err
	}
	return nil
}
