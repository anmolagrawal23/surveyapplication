package main

import (
	"log"
	"net"

	surveyservice "github.com/anmolagrawal23/surveyapplication/proto"
	"github.com/anmolagrawal23/surveyapplication/server"
	"google.golang.org/grpc"
)

func main() {

	log.Println("Starting a grpc server on port 7777")

	lis, err := net.Listen("tcp", ":7777")
	if err != nil {
		log.Fatal("Error listening on port 7777")
	}
	// Create a gRPC server
	gRPCServer := grpc.NewServer()

	// Create a server object of the type we created in server.go
	s := &server.Server{}

	surveyservice.RegisterSurveyServiceServer(gRPCServer, s)
	log.Println(gRPCServer.Serve(lis))
}
