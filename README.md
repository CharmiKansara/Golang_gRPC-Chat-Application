# Chat Application with gRPC in Golang

### Commands:

To generate chat.pb.go - 

``` protoc -I proto  proto/chat.proto  --go_out=plugins=grpc:. ```

For server and client

``` 1. go run server.go ```

Add one or more clients

``` 2. go run client/client.go ```

