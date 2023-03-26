package main

import (
	"flag"
	"fmt"
	"github.com/flonja/sinkingchat/chat"
	"golang.org/x/exp/slices"
)

const LinusTechTips = "/live/5c13f3c006f1be15e08e05c0"

func main() {
	token := flag.String("token", "", "Sails token for socket authentication")
	flag.Parse()

	socket, err := chat.NewFloatplaneChatSocket(LinusTechTips, *token)
	if err != nil {
		panic(err)
	}
	defer func(socket *chat.FloatplaneChatSocket) {
		err := socket.Close()
		if err != nil {
			panic(err)
		}
	}(socket)

	emotes, _ := socket.Emotes()
	fmt.Printf("%#v\n", emotes)
	go func() {
		users, err := socket.Users()
		if err != nil {
			panic(err)
		}
		fmt.Printf("%#v\n", users)
	}()

	if err = socket.Listen(func(message *chat.ResponseRoomMessage) {
		if slices.Contains(message.Mentions(), socket.Username()) && message.UserGuid != socket.Guid() {
			if err = socket.SendMessage(fmt.Sprintf("%v mentioned me in their message!", message.Username)); err != nil {
				panic(err)
			}
		}

		fmt.Printf("%v (self? %v): %v\n", message.Username, message.UserGuid == socket.Guid(), message.Message)
	}); err != nil {
		panic(err)
	}
	//time.Sleep(time.Second) // timeout, since it would otherwise recognise the next message as spam
	//if err = socket.SendMessage("message from golang! :D"); err != nil {
	//	panic(err)
	//}

	// Block the program from exiting
	select {}
}
