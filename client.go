package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
	"time"
)

type Msg struct {
	From string
	Body string
	Kind string
}

var username string
var client *rpc.Client

func pollMessages() {
	for {
		var msgs []Msg
		err := client.Call("ChatService.GetMessages", &username, &msgs)
		if err == nil {
			for _, m := range msgs {

				fmt.Printf("\r[%s] %s\n> ", m.From, m.Body)
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func main() {
	var err error
	client, err = rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Connection error:", err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your username: ")
	username, _ = reader.ReadString('\n')
	username = strings.TrimSpace(username)

	var success bool
	err = client.Call("ChatService.Join", &username, &success)
	if err != nil || !success {
		fmt.Println("Failed to join. Username might be taken or server is down.")
		return
	}

	fmt.Printf("Welcome %s! Type /exit to leave.\n", username)
	fmt.Println("-------------------------------------------------")

	go pollMessages()

	for {
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		text = strings.TrimSpace(text)

		if text == "" {
			continue
		}

		if strings.ToLower(text) == "/exit" {
			client.Call("ChatService.Leave", &username, &success)
			fmt.Println("Exiting...")
			break
		}

		msg := Msg{
			From: username,
			Body: text,
			Kind: "msg",
		}

		client.Call("ChatService.SendMessage", &msg, &success)
	}
}
