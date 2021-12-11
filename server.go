package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
	lpb "grpchat/proto"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type ChatRoom struct {
	id   [32]byte
	name string
}

type ChatUser struct {
	username string
	id       int64
}

var (
	grpcLog         glog.LoggerV2
	serverPort      = flag.Int("port", 50051, "The server port")
	chatRooms       map[string]ChatRoom
	users           map[string]ChatUser
	users2ChatRooms map[string][]string
)

func init() {
	chatRooms = make(map[string]ChatRoom)
	users = make(map[string]ChatUser)
	users2ChatRooms = make(map[string][]string)
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
	chatRooms = make(map[string]ChatRoom)
}

type Connection struct {
	stream lpb.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error
}

type Server struct {
	lpb.UnimplementedBroadcastServer
	Connections []*Connection
}

func (s *Server) GetChatRooms(_ context.Context, req *lpb.User) (*lpb.GetChatResp, error) {
	var chatRoomsForUser []string
	if _, userExists := users[req.Name]; userExists != true {
		//timestamp := time.Now()
		for _, chatName := range users2ChatRooms[req.Name] {
			chatRoomsForUser = append(chatRoomsForUser, chatName)
		}
	}
	return &lpb.GetChatResp{ChatNames: chatRoomsForUser}, nil
}

func (s *Server) CreateChatRoom(_ context.Context, req *lpb.CreateChatReq) (*lpb.GetChatResp, error) {
	chatIdStr := make([]string, 0)
	if _, roomExists := chatRooms[req.ChatName]; roomExists != true {
		timestamp := time.Now()
		id := sha256.Sum256([]byte(timestamp.String() + req.ChatName))

		chatRooms[req.ChatName] = ChatRoom{id, req.ChatName}
		users2ChatRooms[req.User.Name] = append(users2ChatRooms[req.User.Name], req.ChatName)
		chatIdStr = append(chatIdStr, req.ChatName)
	}
	return &lpb.GetChatResp{ChatNames: chatIdStr}, nil
}

func (s *Server) CreateStream(protoConn *lpb.Connect, stream lpb.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream, protoConn.User.Id, true, make(chan error),
	}
	fmt.Printf("CreateStream from user: %v", protoConn.User.Name)
	s.Connections = append(s.Connections, conn)
	return <-conn.error
}

func (s *Server) BroadcastMessage(_ context.Context, msg *lpb.Message) (*lpb.Close, error) {
	wait := sync.WaitGroup{}
	doneChan := make(chan int)

	for _, conn := range s.Connections {
		wait.Add(1)

		go func(msg *lpb.Message, conn *Connection) {
			defer wait.Done()
			if conn.active {
				err := conn.stream.Send(msg)
				grpcLog.Info("Sending msg with user id ", msg.Id, " to: ", conn.stream)
				if err != nil {
					grpcLog.Error("Error with Stream: %s - Error: %v", conn.stream, err)
					conn.active = false
					conn.error <- err
				}
			}
		}(msg, conn)
	}
	go func() {
		wait.Wait()
		close(doneChan)
	}()
	<-doneChan
	return &lpb.Close{}, nil
}

func main() {
	var connections []*Connection
	flag.Parse()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *serverPort))
	if err != nil {
		log.Fatalf("error creating the server %v", err)
	}
	grpcServer := grpc.NewServer()
	server := &Server{}
	server.Connections = connections

	lpb.RegisterBroadcastServer(grpcServer, server)
	grpcLog.Info("Starting server at :", *serverPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
