package main

import (
	sso "github.com/SanctusNiccolum/SiriusLingo/gen/go/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	sso.UnimplementedAuthServiceServer
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	sso.RegisterAuthServiceServer(grpcServer, &Server{})
	log.Println("Auth Service running on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
