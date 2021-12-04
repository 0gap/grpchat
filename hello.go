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
	grpcLog     glog.LoggerV2
	server_port = flag.Int("port", 50051, "The server port")
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
	Connection []*Connection
}

func (s *Server) CreateStream(pconn *lpb.Connect, stream lpb.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream, *pconn.User.Id, true, make(chan error),
	}
	s.Connection = append(s.Connection, conn)
	return <-conn.error
}

func (s *Server) BroadcastMessage(ctx context.Context, msg *lpb.Message) (*lpb.Close, error) {
	wait := sync.WaitGroup{}
	done_chan := make(chan int)

	for _, conn := range s.Connection {
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
		close(done_chan)
	}()
	<-done_chan
	return &lpb.Close{}, nil
}

func main() {
	var connections []*Connection
	flag.Parse()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *server_port))
	if err != nil {
		log.Fatalf("error creating the server %v", err)
	}
	grpcServer := grpc.NewServer()
	server := &Server{}
	server.Connection = connections

	lpb.RegisterBroadcastServer(grpcServer, server)
	grpcLog.Info("Starting server at :", *server_port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
