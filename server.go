package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
	lpb "learningGo/proto"
	"log"
	"net"
	"os"
	"sync"
)

var (
	grpcLog    glog.LoggerV2
	serverPort = flag.Int("port", 50051, "The server port")
)

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
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

func (s *Server) CreateStream(protoConn *lpb.Connect, stream lpb.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream, *protoConn.User.Id, true, make(chan error),
	}
	s.Connections = append(s.Connections, conn)
	return <-conn.error
}

func (s *Server) BroadcastMessage(ctx context.Context, msg *lpb.Message) (*lpb.Close, error) {
	wait := sync.WaitGroup{}
	doneChan := make(chan int)

	for _, conn := range s.Connections {
		wait.Add(1)

		go func(msg *lpb.Message, conn *Connection) {
			defer wait.Done()
			if conn.active {
				err := conn.stream.Send(msg)
				grpcLog.Info("Sending msg with user id ", *msg.Id, " to: ", conn.stream)
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
