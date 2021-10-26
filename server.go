package main

import (
	"grpcChittyChatServer/chittychatserver/MiniP2-ChittyChat2/chittychatserver"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen on port 8080: %v", err)
	}

	grpcServer := grpc.NewServer()

	s := chittychatserver.Server{}

	chittychatserver.RegisterServicesServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port 8080: %v", err)
	}
}
