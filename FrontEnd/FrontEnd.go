package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hpcloud/tail"
	"github.com/lottejd/DISYSMP3/Auction"
	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

const (
	CLIENT_PORT             = 8080
	SERVER_PORT             = 5000
	SERVER_LOG_FILE         = "serverLog"
	REPLICA_STATUS_RESPONSE = "Alive and well"
)

type AuctionType struct {
	Bid    int32
	Bidder int32
	done   bool
}

type FrontEndServer struct {
	Auction.UnimplementedAuctionServiceServer
	this               *AuctionType
	startTime          *time.Time
	replicaServerPorts map[int]bool
}

func main() {
	proxyAuction := AuctionType{-1, -1, false}
	replicaServerMap := make(map[int]bool)
	startTime := time.Now()
	feServer := FrontEndServer{this: &proxyAuction, startTime: &startTime, replicaServerPorts: replicaServerMap}

	go listen(&feServer)
	for {
		feServer.FindReplicasAndAddThemToMap()
	}
}

func (feServer *FrontEndServer) Bid(ctx context.Context, message *Auction.BidRequest) (*Auction.BidResponse, error) {
	currentBid, status := feServer.GetHighestBidFromReplicas() // implement
	fmt.Println(status)

	newBid := message.GetAmount()
	if currentBid.Bid < newBid && feServer.EvalAuctionStillGoing(time.Now()) {
		feServer.UpdateBid(newBid, message.GetClientId())
		return &Auction.BidResponse{Success: true}, nil
	}
	return &Auction.BidResponse{Success: false}, nil

}

func (feServer *FrontEndServer) Result(ctx context.Context, message *Auction.ResultRequest) (*Auction.ResultResponse, error) {
	currentBid, status := feServer.GetHighestBidFromReplicas() // implement
	status = "Succes still going"
	err := errors.New("couldn't find any bids on the running auction")
	var response Auction.ResultResponse
	switch status {
	case "Failed":
		return &Auction.ResultResponse{BidderID: -1, HighestBid: -1, Done: false}, err
	case "Succes still going":
		response = Auction.ResultResponse{BidderID: currentBid.Bidder, HighestBid: currentBid.Bid, Done: false}
	case "Succes done":
		response = Auction.ResultResponse{BidderID: currentBid.Bidder, HighestBid: currentBid.Bid, Done: true}
	}
	return &response, nil

}

func (feServer *FrontEndServer) UpdateBid(bid int32, bidderId int32) {

	newAuction := AuctionType{Bid: bid, Bidder: bidderId, done: false}
	msg := feServer.UpdateAllReplicas(newAuction) // implement
	feServer.this = &newAuction
	fmt.Println(msg)
}

func (feServer *FrontEndServer) CreateAuction(logMsg string) (AuctionType, int) {
	splitMsg := strings.Fields(logMsg)
	var bid, bidder, replicaPort int
	for i, subMsg := range splitMsg {
		switch subMsg {
		case "HighestBid:":
			fmt.Println(splitMsg[i+1])
			bid, _ = strconv.Atoi(splitMsg[i+1])
		case "id:":
			fmt.Println(splitMsg[i+1])
			bidder, _ = strconv.Atoi(splitMsg[i+1])
		case "port:":
			fmt.Println(splitMsg[i+1])
			replicaPort, _ = strconv.Atoi(splitMsg[i+1])
			break
		}
	}
	return AuctionType{Bid: int32(bid), Bidder: int32(bidder), done: feServer.EvalAuctionStillGoing(time.Now())}, replicaPort
}

func (feServer *FrontEndServer) GetHighestBidFromReplicas() (AuctionType, string) {
	// implement
	latestLogs := feServer.ReadFromLog()
	latestAuctionsFromLogs := make([]AuctionType, 10)
	for _, Msg := range latestLogs {
		auction, replicaPort := feServer.CreateAuction(Msg)
		latestAuctionsFromLogs = append(latestAuctionsFromLogs, auction)
		fmt.Println(replicaPort)
	}

	// splitLatestMsg := strings.Fields(latestLogs[len(latestLogs)-1])
	for _, auctions := range latestAuctionsFromLogs {
		fmt.Println(auctions)
	}
	fmt.Println(len(latestAuctionsFromLogs))

	// get bid from all replicas maybe use timestamps or just r = w = n-1/2
	// the value repeating the most or quorum wins
	return *feServer.this, "succesfully got the bid from client"
}

func (feServer *FrontEndServer) ReadFromLog() []string {
	replicaMsgs := make([]string, len(feServer.replicaServerPorts))
	err := os.Chdir("../Server") // læs fra server directory
	if err != nil {
		fmt.Println(err)
	}
	for replica, alive := range feServer.replicaServerPorts {
		if alive {
			t, err := tail.TailFile(fmt.Sprintf("%s%v", SERVER_LOG_FILE, replica), tail.Config{Follow: true})
			if err != nil {
				fmt.Printf("Error: %v\n", err)

			}

			var latestMsg string
			go t.StopAtEOF() // stop after last line has been read

			for line := range t.Lines {
				latestMsg = line.Text
			}
			latestMsgAndPort := fmt.Sprintf("%s  replica port: %v", latestMsg, replica)
			replicaMsgs = append(replicaMsgs, latestMsgAndPort)
		}
	}
	return replicaMsgs

}

func (feServer *FrontEndServer) UpdateAllReplicas(auction AuctionType) string {
	var failedWrites int
	ctx := context.Background()
	request := Replica.Auction{Bid: auction.Bid, BidId: auction.Bidder}

	for port, alive := range feServer.replicaServerPorts {
		if alive {
			conn, status := Connect(port)
			if status == "succes" {
				replicaClient, status := ConnectToReplicaClient(conn)
				response, err := replicaClient.WriteToLog(ctx, &request)
				if err != nil {
					fmt.Println(err)
				} else if response.GetMsg() != "succes" || status == "failed" {
					failedWrites++
				}
				conn.Close()
			}
		}
	}
	if failedWrites/2 < len(feServer.replicaServerPorts) {
		return "succesfully updated replica logs"
	}
	return "failed to update more than half of the replicas"
}

// find replicas and add
func (feServer *FrontEndServer) FindReplicasAndAddThemToMap() {
	ctx := context.Background()
	counter := 0
	i := 0
	for counter <= 3 {
		port := SERVER_PORT + i
		conn, status := Connect(port)
		i++

		if status == "succes" {
			replicaClient, status := ConnectToReplicaClient(conn)
			response, err := replicaClient.CheckStatus(ctx, &Replica.EmptyRequest{})
			if err != nil || status == "failed" {
				feServer.replicaServerPorts[port] = false
			} else if response.GetMsg() == REPLICA_STATUS_RESPONSE {
				feServer.replicaServerPorts[port] = true
				// fmt.Printf("found port: %v\n", port)
			}
		}
		counter++
	}
	ctx.Done()
}

func (feServer *FrontEndServer) EvalAuctionStillGoing(current time.Time) bool {

	fmt.Println(feServer.startTime.Format(time.Layout))
	fmt.Println(current.Format(time.Layout))
	return true // implement
}

// grcp server setup
func listen(feServer *FrontEndServer) {
	address := fmt.Sprintf("localhost:%v", CLIENT_PORT)
	lis, _ := net.Listen("tcp", address)
	defer lis.Close()

	grcpServer := grpc.NewServer()
	Auction.RegisterAuctionServiceServer(grcpServer, feServer)
	if err := grcpServer.Serve(lis); err != nil {
		fmt.Printf("Failed to serve AuctionService on port %v\n", CLIENT_PORT)
	}
}

// connect to a replica with at max 3 retries
func ConnectToReplicaClient(conn *grpc.ClientConn) (Replica.ReplicaServiceClient, string) {
	ctx := context.Background()

	client := Replica.NewReplicaServiceClient(conn)
	ack, err := client.CheckStatus(ctx, &Replica.EmptyRequest{})
	if err != nil || ack.GetMsg() == "" {
		return client, "failed"
	}

	ctx.Done()
	return client, "succes"
}

// connect to address
func Connect(port int) (*grpc.ClientConn, string) {
	address := fmt.Sprintf("localhost:%v", port)
	var status string
	var conn *grpc.ClientConn
	for i := 0; i < 3; i++ {
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			status = "failed"
			time.Sleep(time.Millisecond * 250)
			continue
		} else {
			return conn, "succes"
		}
	}
	return conn, status
}
