package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/lottejd/DISYSMP3/Auction"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:8080"
	logFileName = "logfile"
)

var (
	id int32
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
	client := Auction.NewAuctionServiceClient(conn)
	ctx := context.Background()

	go listenForInput(&client, ctx)
}

func Bid(client *Auction.AuctionServiceClient, amount int) {

}

func Result(client *Auction.AuctionServiceClient) {

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

func listenForInput(client *Auction.AuctionServiceClient, ctx context.Context) {
	for {
		// to ensure "enter" has been hit before publishing - skud ud til Mie
		reader, err := bufio.NewReader(os.Stdin).ReadString('\n')
		// remove newline windows format "\r\n"
		input := strings.TrimSuffix(reader, "\r\n")
		if err != nil {
			Logger("bad bufio input", logFileName)
		}
		if len(input) > 0 {
			switch input {
			case "bid":
				fmt.Print("How much would you like to bid?")
				reader, err := bufio.NewReader(os.Stdin).ReadString('\n')
				input := strings.TrimSuffix(reader, "\r\n")
				if err != nil {
					Logger("bad bufio input", logFileName)
				}
				bid, err := strconv.Atoi(input)
				Bid(client, bid)
				break

			case "result":
				Result(client)
				break
			}
		}
	}
}
