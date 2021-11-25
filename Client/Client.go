package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/lottejd/DISYSMP3/Auction"
	"google.golang.org/grpc"
)

const (
	ADDRESS       = "localhost:8080"
	LOG_FILE_NAME = "logfile"
)

type User struct {
	userId int32
}

func main() {
	// init
	// Set up a connection to the server.
	ctx := context.Background()
	conn, err := grpc.Dial(ADDRESS, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// create client
	var id int32

	fmt.Println("Choose an integer as id:")
	fmt.Scanln(&id)
	user := User{userId: int32(id)}
	client := Auction.NewAuctionServiceClient(conn)

	go user.listenForInput(client, ctx)

	for {
	}
}

// client side
func (u *User) Bid(client Auction.AuctionServiceClient, ctx context.Context, amount int32) {
	request := &Auction.BidRequest{Amount: amount, ClientId: u.userId}
	response, err := client.Bid(ctx, request)
	if response.Success {
		fmt.Println("Bid was successful")
	} else if err != nil {
		fmt.Errorf(err.Error())
	} else {
		fmt.Println("Bid was not successful. Check if your bid was an int, and higher than the current bid")
	}
}

func (u *User) Result(client Auction.AuctionServiceClient, ctx context.Context) {
	request := &Auction.ResultRequest{}
	response, err := client.Result(ctx, request)
	if response.GetDone() {
		fmt.Printf("The auction is finished. Bidder with id: %v won the auction, with bid: %v", response.GetBidderID(), response.GetHighestBid())
	} else if err == nil {
		fmt.Printf("The auction is still going. Current highest bid comes from bidder: %v, who is bidding: %v", response.GetBidderID(), response.GetHighestBid())
	} else {
		fmt.Errorf(err.Error())
	}

}

func (u *User) listenForInput(client Auction.AuctionServiceClient, ctx context.Context) {
	for {
		var input string
		fmt.Scanln(&input)
		if len(input) > 0 {
			switch input {
			case "bid":
				fmt.Println("How much would you like to bid?")
				fmt.Scanln(&input)
				bid, err := strconv.Atoi(input)
				if err != nil {
					fmt.Errorf("Error: %v", err)
				} else {
					u.Bid(client, ctx, int32(bid))
				}
				break

			case "result":
				u.Result(client, ctx)
				break
			}
		}
	}
}

// below is not used
//*Auction.AuctionServiceClient implement
func reconnect(client Auction.AuctionServiceClient) {
	var tempConnected bool
	var temp Auction.AuctionServiceClient
	for !tempConnected {
		temp, IsConnected := ConnectAsAuctionClient()
		tempConnected = IsConnected
		client = temp
	}
	client = temp
}

func ConnectAsAuctionClient() (Auction.AuctionServiceClient, bool) {
	conn, err := grpc.Dial(ADDRESS, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// create client
	var client Auction.AuctionServiceClient
	IsConnected := false
	client = Auction.NewAuctionServiceClient(conn)
	request := Auction.ResultRequest{}
	response, err := client.Result(context.Background(), &request)
	if err != nil {
		response.Reset() // idk man need to use this variable and cant set it to  _, err := client.REsult(sad)
		IsConnected = true
	}
	return client, IsConnected
}
