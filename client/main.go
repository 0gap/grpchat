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
	"sync"
	"time"
)

var (
	client lpb.BroadcastClient
	wait   *sync.WaitGroup
)

func init() {
	wait = &sync.WaitGroup{}
}

func connect(user *lpb.User) error {
	var streamerror error
	active := true
	stream, err := client.CreateStream(context.Background(), &lpb.Connect{
		User: user, Active: &active,
	})
	if err != nil {
		return fmt.Errorf("connaction failed: %v", err)
	}

	wait.Add(1)

	go func(str lpb.Broadcast_CreateStreamClient) {
		defer wait.Done()

		for {
			msg, err := str.Recv()
			if err != nil {
				streamerror = fmt.Errorf("error receiving msg: %v", err)
				break
			}
			fmt.Printf("Received %v from user %s\n", *msg.Content, *msg.Id)
		}
	}(stream)
	return streamerror
}

func main() {
	timestamp := time.Now()
	doneChan := make(chan int)

	name := flag.String("n", "panos", "The name of the user sending")
	flag.Parse()

	id := sha256.Sum256([]byte(timestamp.String() + *name))

	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldn't connect to the service: %v", err)
	}

	client = lpb.NewBroadcastClient(conn)
	hexId := hex.EncodeToString(id[:])
	user := &lpb.User{
		Id:   &hexId,
		Name: name,
	}
	err = connect(user)
	if err != nil {
		return
	}

	wait.Add(1)

	go func() {
		defer wait.Done()
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			content := scanner.Text()
			t := timestamp.String()
			msg := &lpb.
				Message{Id: user.Id, Content: &content, Timestamp: &t}
			_, err := client.BroadcastMessage(context.Background(), msg)
			if err != nil {
				fmt.Printf("Error Sending message: %v", err)
				break
			}
		}
	}()

	go func() {
		wait.Wait()
		close(doneChan)
	}()

	<-doneChan
}
