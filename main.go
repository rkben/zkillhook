package main

import (
	"github.com/sacOO7/gowebsocket"
	"log"
	"os"
	"os/signal"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"bytes"
)

const zkillws string = "wss://zkillboard.com:2096"
var discordurl string

type details struct {
	discordurl string
	id string
}

func zConnect(socket gowebsocket.Socket, p details) {
	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Printf("Having a chat with zKillboard\nWebhook: %s\nID: %s", p.discordurl, p.id)
		socket.SendText(fmt.Sprintf("{\"action\":\"sub\",\"channel\":\"%s\"}", p.id))
	}
	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		log.Println("RECV ", message)
		url := handlemsg(message)
		go postDiscord(url, p.discordurl)
	}
	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Stopped chatting with zKillboard")
		return
	}

	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Println(err)
	}
	socket.Connect()
}

func handlemsg(message string) string{
	var msg map[string]string
	json.Unmarshal([]byte(message), &msg)
	return msg["url"]
}

func postDiscord(zkillurl string, discordurl string) {
	log.Println(zkillurl)
	req, _ := json.Marshal(map[string]string{
		"username": "zKillboard",
		"icon_url": "https://image.eveonline.com/Render/670_512.png",
		"text": zkillurl,
	})
	resp, _ := http.Post(discordurl, "application/json", bytes.NewBuffer(req))
	resp.Body.Close()
}

func main() {
	params := new(details)
	flag.StringVar(&params.discordurl, "url", "", "Discord Webhook URL")
	flag.StringVar(&params.id, "filter", "killstream", "zKillboard filter (see zKillboard Websocket docs)")
	flag.Parse()
	if params.discordurl != "" {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		socket := gowebsocket.New(zkillws)
		zConnect(socket, *params)
		for {
			select {
			case <-interrupt:
				log.Println("got interrupt")
				socket.Close()
				return
			}
		}

	}
	log.Println("Need a webhook URL")

}
