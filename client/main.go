package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	lpb "grpchat/proto"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type ChatRoom struct {
	id   [32]byte
	name string
}

var (
	client           lpb.BroadcastClient
	wait             *sync.WaitGroup
	clientChatRooms  []ChatRoom
	serverPort       *int
	name             *string
	selectedChatRoom string
)

func init() {
	wait = &sync.WaitGroup{}
	clientChatRooms = make([]ChatRoom, 0)
	serverPort = flag.Int("p", 50051, "The server port")
	name = flag.String("u", "panos", "The name of the user sending")
	selectedChatRoom = ""
}

func connect(user *lpb.User) error {
	var streamError error
	stream, err := client.CreateStream(context.Background(), &lpb.Connect{
		User: user, Active: true,
	})
	if err != nil {
		return fmt.Errorf("connaction failed: %v", err)
	}
	roomList, err := client.GetChatRooms(context.Background(), user)
	if err != nil {
		return fmt.Errorf("Getting chat rooms failed")
	}
	fmt.Println("Chat rooms for user: %s", user.Name)
	for _, chatName := range roomList.ChatNames {
		timestamp := time.Now()
		id := sha256.Sum256([]byte(timestamp.String() + chatName))
		clientChatRooms = append(clientChatRooms, ChatRoom{id, chatName})
		fmt.Println("\t%s", chatName)
		selectedChatRoom = chatName
	}
	fmt.Println("Total rooms: %d", len(clientChatRooms))

	wait.Add(1)

	go func(str lpb.Broadcast_CreateStreamClient) {
		defer wait.Done()
		for {
			msg, err := str.Recv()
			if err != nil {
				streamError = fmt.Errorf("error receiving msg: %v", err)
				break
			}
			fmt.Printf("Received %v from user %s\n", msg.Content, msg.Id)
		}
	}(stream)
	return streamError
}

type Command struct {
	cmdLiteral string
	cmdArg     string
}

func executeCommand(cmdChan chan Command, user *lpb.User) {
	var cmd Command
	for true {
		cmd = <-cmdChan
		// create channel
		if cmd.cmdLiteral == "cc" {
			newChatName := strings.Split(cmd.cmdArg, " ")[0]
			createReq := &lpb.CreateChatReq{
				ChatName: newChatName,
				User:     user,
			}
			_, err := client.CreateChatRoom(context.Background(), createReq)
			if err != nil {
				log.Printf("Error creating chat room: %v\n", err)
				break
			}
			selectedChatRoom = newChatName
		}
		// send msg
		if cmd.cmdLiteral == "s" {
			if selectedChatRoom == "" {
				log.Printf("You need to be in a chat room. Select(sc) or Create(cc) one.")
				continue
			}
			timestamp := time.Now()
			t := timestamp.String()
			msg := &lpb.
				Message{Id: user.Id, Content: cmd.cmdArg, Timestamp: t, ChatName: selectedChatRoom}
			_, err := client.BroadcastMessage(context.Background(), msg)
			if err != nil {
				log.Printf("Error Sending message: %v\n", err)
				break
			}
		}
		// select channel
		if cmd.cmdLiteral == "sc" {
			selectedChatRoom = cmd.cmdArg
			log.Printf("Selected chat room %s\n", selectedChatRoom)
		}
	}
}

func main() {
	timestamp := time.Now()
	doneChan := make(chan int)

	flag.Parse()

	id := sha256.Sum256([]byte(timestamp.String() + *name))
	// 	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", serverPort), grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldn't connect to the service: %v", err)
	}

	client = lpb.NewBroadcastClient(conn)
	hexId := hex.EncodeToString(id[:])
	user := &lpb.User{
		Id:   hexId,
		Name: *name,
	}
	err = connect(user)
	if err != nil {
		return
	}

	cmdChan := make(chan Command)

	wait.Add(1)
	go executeCommand(cmdChan, user)

	wait.Add(1)
	go func() {
		defer wait.Done()
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			content := strings.SplitN(scanner.Text(), " ", 2)
			if len(content) != 2 {
				log.Printf("Erroneous command: %v", err)
				continue
			}
			cmd := Command{cmdLiteral: content[0], cmdArg: content[1]}
			cmdChan <- cmd
		}
	}()

	go func() {
		wait.Wait()
		close(doneChan)
	}()

	<-doneChan
}
