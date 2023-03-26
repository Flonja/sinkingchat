package chat

import "regexp"

var mentionRegex = regexp.MustCompile(`(?m)@([\w-]+)`)

type ResponseRoomMessage struct {
	Id       string  `json:"id"`
	UserGuid string  `json:"userGUID"`
	Username string  `json:"username"`
	Channel  string  `json:"channel"`
	Message  string  `json:"message"`
	UserType string  `json:"userType"`
	Emotes   []Emote `json:"emotes"`
	Success  bool    `json:"success"`
}

func (m *ResponseRoomMessage) Mentions() []string {
	matches := mentionRegex.FindStringSubmatch(m.Message)
	if len(matches) == 0 {
		return nil
	}

	return matches[1:]
}

type ResponseUserList struct {
	Pilots     []string `json:"pilots"`
	Passengers []string `json:"passengers"`
}

type ResponseJoinRoom struct {
	Success bool     `json:"success"`
	Emotes  *[]Emote `json:"emotes"`
}

type Emote struct {
	Code  string `json:"code"`
	Image string `json:"image"`
}
