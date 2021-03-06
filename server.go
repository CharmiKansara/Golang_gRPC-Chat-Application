package main

import (
	"chatapp/proto"
	"log"
	"net"
	"os"

	"sync"

	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"

	//"google.golang.org/protobuf/proto"

	glog "google.golang.org/grpc/grpclog"
)

type Connection struct {
	stream proto.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error
}

type Server struct {
	Connection []*Connection
}

func (s *Server) CreateStream(pconn *proto.Connect, stream proto.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream: stream,
		id:     pconn.User.Id,
		active: true,
		error:  make(chan error),
	}

	s.Connection = append(s.Connection, conn)
	return <-conn.error
}

func (s *Server) BroadcastMessage(ctx context.Context, msg *proto.Message) (*proto.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range s.Connection {
		wait.Add(1)
		go func(msg *proto.Message, conn *Connection) {
			defer wait.Done()
			if conn.active {
				err := conn.stream.Send(msg)
				if err != nil {
					grpcLog.Errorf("error in stream: %v", conn.stream)
					conn.active = false
					conn.error <- err
				}
			}
		}(msg, conn)
	}
	go func() {
		wait.Wait()
		close(done)
	}()
	<-done

	return &proto.Close{}, nil
}

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

func main() {
	var connections []*Connection

	server := &Server{connections}

	grpcServer := grpc.NewServer()
	listner, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error in creating server %v", err)
	}

	grpcLog.Info("Server started at 8080")

	proto.RegisterBroadcastServer(grpcServer, server)
	grpcServer.Serve(listner)
}
