package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"
	"strconv"
	"fmt"

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
	id              int32
	primary         bool
	port            int32
	nextReplicaPort int32
	highestBid int32
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
	if(message.Amount > s.highestBid) {
		s.updateBid(message.Amount)
		return &Auction.BidResponse{Success: true},nil
	}
	return &Auction.BidResponse{Success: true},nil
}

func (s *Server) Result(ctx context.Context, message *Auction.ResultRequest) (*Auction.ResultResponse, error) {
	return &Auction.ResultResponse{AuctionID: s.highestBid},nil
}

func (s *Server) askNextReplica() {
	ctx := context.Background()
	address := fmt.Sprintf("localhost:%v", s.nextReplicaPort)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to: %s", strconv.FormatInt(int64(s.port),10))
	}

	nextRep := Auction.NewAuctionServiceClient(conn)

	if _, err := nextRep.askNextReplica(); err != nil {
		log.Println(err)
	} else {
		log.Println("No errors")
	}
}

func (s *Server) updateBid(bid int32) {
	s.highestBid = bid
	//reportToPrimary()
}

func StartAuction() {

}

func EndAuction() {

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