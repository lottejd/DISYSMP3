package main

import (
	"context"
	"github.com/lottejd/DISYSMP3/Auction"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

const (
	port          = ":8080"
	serverLogFile = "serverLog"
)

type Server struct {
	Auction.UnimplementedAuctionServiceServer
}

func main() {

	//init
	grpcServer := grpc.NewServer()

	//setup listen on port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// start the service / server on the specific port
	Auction.RegisterAuctionServiceServer(grpcServer, &Server{})
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port %s  %v", port, err)
	}

}

func (s *Server) Bid(ctx context.Context, message *Auction.BidRequest) (*Auction.BidResponse, error) {

}

func Logger(message string, logFileName string) {
	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(message)
}
