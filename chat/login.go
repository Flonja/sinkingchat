package chat

import (
	"fmt"
	"net"
	"net/url"
	"time"
	"github.com/tebeka/selenium"
)

func pickUnusedPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err := l.Close(); err != nil {
		return 0, err
	}
	return port, nil
}

func Login() string {
	const chromeDriverPath = "/usr/local/bin/chromedriver"


	// Set up Selenium and Chromedriver
	port, err := pickUnusedPort()
	if err != nil {
		panic(err)
	}
	service, err := selenium.NewChromeDriverService(chromeDriverPath, port)
	if err != nil {
		panic(err)
	}
	defer service.Stop()

	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	/*
	if err := wd.Get("https://chat.floatplane.com/__getcookie"); err != nil {
		panic(err)
	}
	*/

	// Open login page for user to enter credentials
	if err := wd.Get("http://www.floatplane.com/login"); err != nil {
		panic(err)
	}

	// Wait for user to log in
	var current_url string
	for {
		current_url, err = wd.CurrentURL()
		if err != nil {
			panic(err)
		}
		if current_url != "https://www.floatplane.com/login" {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	cookie, err := wd.GetCookie("sails.sid")
	if err != nil {
		panic(err)
	}

	token, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		panic(err)
	}

	return token
}
