package chat

type Request struct {
	Method  string            `json:"method"`
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Data    map[string]any    `json:"data"`
}

type Response struct {
	Body       any               `json:"body"`
	Headers    map[string]string `json:"headers"`
	StatusCode int               `json:"statusCode"`
}

const (
	JoinRoom    = "/RadioMessage/joinLivestreamRadioFrequency"
	LeaveRoom   = "/RadioMessage/leaveLivestreamRadioFrequency"
	SendRoomMsg = "/RadioMessage/sendLivestreamRadioChatter/"
	GetUserList = "/RadioMessage/getChatUserList/"
)

func newJoinLivestreamRequest(channel string) *Request {
	return &Request{Method: "get", Url: JoinRoom, Headers: map[string]string{}, Data: map[string]any{
		"channel": channel,
		"message": nil,
	}}
}

func newLeaveLivestreamRequest(channel string) *Request {
	return &Request{Method: "get", Url: LeaveRoom, Headers: map[string]string{}, Data: map[string]any{
		"channel": channel,
		"message": "bye!",
	}}
}

func newSendLivestreamMsgRequest(channel, message string) *Request {
	return &Request{Method: "post", Url: SendRoomMsg, Headers: map[string]string{}, Data: map[string]any{
		"channel": channel,
		"message": message,
	}}
}

func newGetUserListRequest(channel string) *Request {
	return &Request{Method: "get", Url: GetUserList, Headers: map[string]string{}, Data: map[string]any{
		"channel": channel,
	}}
}
