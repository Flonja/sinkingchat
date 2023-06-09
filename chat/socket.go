package chat

import (
	"encoding/json"
	"fmt"
	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"github.com/mitchellh/mapstructure"
	"io"
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

func NewFloatplaneChatSocket(channel string, sailsToken string) (*FloatplaneChatSocket, error) {
	if !strings.HasPrefix(sailsToken, "s:") {
		return nil, fmt.Errorf("sailsToken doesn't start with `s:`, likely not valid")
	}
	requestHeader := http.Header{}
	requestHeader.Set("Origin", "https://www.floatplane.com")
	requestHeader.Set("Cookie", fmt.Sprintf("sails.sid=%v", sailsToken))

	wssUrl := "wss://chat.floatplane.com/socket.io/?__sails_io_sdk_version=0.13.8&__sails_io_sdk_platform=browser&__sails_io_sdk_language=javascript&EIO=3&transport=websocket"
	trnsprt := transport.GetDefaultWebsocketTransport()
	trnsprt.RequestHeader = requestHeader
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

	req, err := http.NewRequest("GET", "https://www.floatplane.com/api/v3/user/self", nil)
	if err != nil {
		return nil, err
	}
	req.Header = requestHeader
	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if b, err := io.ReadAll(resp.Body); err == nil {
		var m map[string]any
		if err = json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		socket.username = m["username"].(string)
		socket.guid = m["id"].(string)
	}

	return socket, nil
}

func (f *FloatplaneChatSocket) Username() string {
	return f.username
}

func (f *FloatplaneChatSocket) Guid() string {
	return f.guid
}

// Emotes gets all current emotes that are available.
func (f *FloatplaneChatSocket) Emotes() ([]Emote, bool) {
	return *f.emotes, f.emotes != nil
}

// Users gets the entire list of users watching/waiting for the stream.
func (f *FloatplaneChatSocket) Users() (*ResponseUserList, error) {
	d := &ResponseUserList{}
	return d, ack(f, newGetUserListRequest(f.channel), d)
}

// SendMessage sends a message to the chat.
func (f *FloatplaneChatSocket) SendMessage(message string) error {
	var out *any = nil
	return ack(f, newSendLivestreamMsgRequest(f.channel, message), out)
}

// SendMessageEmit sends a message to the chat without waiting for it to be acknowledged.
func (f *FloatplaneChatSocket) SendMessageEmit(message string) error {
	return f.c.Emit("post", newSendLivestreamMsgRequest(f.channel, message))
}

// Close formally exits the socket from the room and closes the socket too
func (f *FloatplaneChatSocket) Close() error {
	var out *any = nil
	if err := ack(f, newLeaveLivestreamRequest(f.channel), out); err != nil {
		return err
	}
	f.c.Close()
	return nil
}

// CloseEmit formally exits the socket from the room without waiting for it to be acknowledged and closes the socket too.
func (f *FloatplaneChatSocket) CloseEmit() error {
	if err := f.c.Emit("get", newLeaveLivestreamRequest(f.channel)); err != nil {
		return err
	}
	f.c.Close()
	return nil
}

// Listen is a event listener for any chat message that may get sent while being in the room.
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

	if out == nil {
		// Does not want to decode the response
		return nil
	}
	r := &Response{}
	if err = json.Unmarshal([]byte(resp), r); err != nil {
		return err
	}
	if r.StatusCode < 200 && r.StatusCode >= 300 {
		return fmt.Errorf("unsuccessful status code: %v", resp)
	}
	if err = mapstructure.Decode(r.Body, out); err != nil {
		return err
	}
	return nil
}
