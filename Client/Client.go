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
	ADDRESS        = "localhost:8080"
	LOG_FILE_NAME  = "logfile"
	BID_UNSUCCEFUL = "Bid was not successful. Check if your bid was an int, and higher than the current bid"
	NO_BID_YET     = "There is no accepted bid yet"
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

	user.listenForInput(client, ctx)
}

// client side
func (u *User) Bid(client Auction.AuctionServiceClient, ctx context.Context, amount int32) string {
	request := &Auction.BidRequest{Amount: amount, ClientId: u.userId}
	response, err := client.Bid(ctx, request)
	var reply string
	if response.Success {
		reply = "Bid was successful"
	} else if err != nil {
		fmt.Errorf(err.Error())
	} else {
		response, err := client.Result(ctx, &Auction.ResultRequest{})
		if (response.GetDone() || response.GetHighestBid() == 0) && err == nil {
			reply = u.Result(client, ctx)
		} else {
			reply = BID_UNSUCCEFUL
		}
	}
	return reply
}

func (u *User) Result(client Auction.AuctionServiceClient, ctx context.Context) string {
	request := &Auction.ResultRequest{}
	response, err := client.Result(ctx, request)
	var reply string
	if response.GetHighestBid() != -1 {
		if response.GetDone() {
			reply = fmt.Sprintf("The auction is finished. Bidder with id: %v won the auction, with bid: %v", response.GetBidderID(), response.GetHighestBid())
		} else if err == nil {
			reply = fmt.Sprintf("The auction is still going. Current highest bid comes from bidder: %v, who is bidding: %v", response.GetBidderID(), response.GetHighestBid())
		}
	} else {
		if err != nil {
			fmt.Errorf(err.Error())
		}
		reply = NO_BID_YET
	}
	return reply
}

func (u *User) listenForInput(client Auction.AuctionServiceClient, ctx context.Context) {
	for {
		var input string
		fmt.Scanln(&input)
		if len(input) > 0 {
			switch input {
			case "bid":
				fmt.Println("How much would you like to bid?")
				var bid string
				fmt.Scanln(&bid)
				bidAmount, err := strconv.Atoi(bid)
				if err != nil || len(bid) == 0 {
					fmt.Println("Bad input only integers allowed")
					fmt.Errorf("Error: %v", err)
				} else {
					fmt.Println(u.Bid(client, ctx, int32(bidAmount)))
				}
				break

			case "result":
				fmt.Println(u.Result(client, ctx))
				break
			}
		}
	}
}
