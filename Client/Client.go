package main

import (
	"context"
	"log"
	"os"

	"github.com/lottejd/DISYSMP3/Auction"
	"google.golang.org/grpc"
)

const (
	address = "localhost:8080"
)

func main() {
	// init
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// create client
	auction := Auction.NewAuctionServiceClient(conn)
	ctx := context.Background()
}

func Bid() {

}

func Result() {

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
