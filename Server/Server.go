package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/lottejd/DISYSMP3/Auction"
	"google.golang.org/grpc"
)

const (
	port          = ":8080"
	serverLogFile = "serverLog"
)

var (
	currentAuctionId int32
)

type Server struct {
	Auction.UnimplementedAuctionServiceServer
}

func main() {

	//init
	grpcServer := grpc.NewServer()
	currentAuctionId = 0

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

	for {
		StartAuction()
		time.Sleep(time.Minute * 5)
		EndAuction()
	}

}

func (s *Server) Bid(ctx context.Context, message *Auction.BidRequest) (*Auction.BidResponse, error) {

}

func (s *Server) Result(ctx context.Context, message *Auction.ResultRequest) (*Auction.ResultResponse, error) {

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

func StartAuction() {
	currentAuctionId++
}

func EndAuction() {

}