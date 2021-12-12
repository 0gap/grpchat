Testing out grpc and Go.

A simple chat client server implementation. My baby steps with Go.

### Server

In project root folder, run `go build` and run executable.
Executable flags:
- `-p`: port to which server is listening

### Client

In `./client` run `go build` and run executable.
Executable flags:
- `-p <port No>`: server port
- `-u <username>`: username with which you connect

Client is getting terminal input and recognizes 3 commands:
- `cc <chat room name>`: Create chat room
- `sc <chat room name>`: Select chat room
- `s <message>`: Send message to last selected chat

When a user has selected a chat room and sends a msg, server will send it only towards the user connections which have accessed the said chat room.

### References

- https://steemit.com/utopian-io/@tensor/building-a-chat-app-with-docker-and-grpc?utm_source=pocket_mylist
- https://grpc.io/docs/languages/go/basics/
- https://medium.com/@viethapascal/golang-grpc-part-2-simple-chat-application-with-grpc-ef6a6c0eea32