package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	auction "github.com/lottejd/DISYSMP3/Auction"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:8080"
	logFileName = "logfile"
)

var (
	clientID int32
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
	client := auction.NewAuctionServiceClient(conn)
	ctx := context.Background()

	go listenForInput(client, ctx)
	for {
		time.Sleep(time.Millisecond * 250)
		fmt.Scanln()
	}

}

func Bid(client auction.AuctionServiceClient, ctx context.Context, amount int32) {
	request := &auction.BidRequest{Amount: amount, ClientId: clientID}
	response, err := client.Bid(ctx, request)
	if response.Success {
		fmt.Println("Bid was successful")
	} else if err != nil {
		fmt.Errorf(err.Error())
	} else {
		fmt.Println("Bid was not successful. Check if your bid was an int, and higher than the current bid")
	}
}

func Result(client auction.AuctionServiceClient, ctx context.Context) {
	request := &auction.ResultRequest{}
	response, err := client.Result(ctx, request)
	if response.Done {
		fmt.Printf("The auction is finished. Bidder with id: %v won the auction, with bid: %v", response.BidderID, response.HighestBid)
	} else if err == nil {
		fmt.Printf("The auction is still going. Current highest bid comes from bidder: %v, who is bidding: %v", response.BidderID, response.HighestBid)
	} else {
		fmt.Errorf(err.Error())
	}

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

func listenForInput(client auction.AuctionServiceClient, ctx context.Context) {
	for {
		reader, err := bufio.NewReader(os.Stdin).ReadString('\n')
		input := strings.TrimSuffix(reader, "\r\n")
		if err != nil {
			Logger("bad bufio input", logFileName)
		}
		if len(input) > 0 {
			switch input {
			case "bid":
				fmt.Println("How much would you like to bid?")
				reader, err := bufio.NewReader(os.Stdin).ReadString('\n')
				input := strings.TrimSuffix(reader, "\r\n")
				if err != nil {
					Logger("bad bufio input", logFileName)
				}
				bid, err := strconv.Atoi(input)
				if err != nil {
					log.Fatal(err)
					fmt.Print("Error: Check logs")
				} else {
					Bid(client, ctx, int32(bid))
				}
				break

			case "result":
				Result(client, ctx)
				break
			}
		}
	}
}
