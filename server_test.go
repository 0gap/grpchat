package main

import (
	"context"
	lpb "grpchat/proto"
	"testing"
)

func TestServerChatCreation(t *testing.T) {
	server := &Server{}
	user := &lpb.User{
		Id:   "userId",
		Name: "userName",
	}
	req := &lpb.CreateChatReq{ChatName: "test_chat", User: user}
	if resp, err := server.CreateChatRoom(context.Background(), req); err != nil || len(resp.ChatNames) != 1 || resp.ChatNames[0] != "test_chat" {
		t.Fatalf("Creation of chat failed")
	}
}

func TestServer_GetChatRooms(t *testing.T) {
	// First create 2 chat rooms
	server := &Server{}
	user := &lpb.User{
		Id:   "userId_1",
		Name: "userName-1'",
	}
	req := &lpb.CreateChatReq{ChatName: "test_chat_1", User: user}
	if resp, err := server.CreateChatRoom(context.Background(), req); err != nil || len(resp.ChatNames) != 1 || resp.ChatNames[0] != "test_chat_1" {
		t.Fatalf("Creation of chat failed")
	}
	req = &lpb.CreateChatReq{ChatName: "test_chat_2", User: user}
	if resp, err := server.CreateChatRoom(context.Background(), req); err != nil || len(resp.ChatNames) != 1 || resp.ChatNames[0] != "test_chat_2" {
		t.Fatalf("Creation of chat failed")
	}
	if resp, err := server.GetChatRooms(context.Background(), user); err != nil || len(resp.ChatNames) != 2 || resp.ChatNames[0] != "test_chat_1" || resp.ChatNames[1] != "test_chat_2" {
		t.Fatalf("Creation of chat failed")
	}
}
