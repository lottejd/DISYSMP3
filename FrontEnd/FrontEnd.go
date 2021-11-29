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
	CLIENT_PORT              = 8080
	SERVER_PORT              = 5000
	SERVER_LOG_FILE          = "serverLog"
	REPLICA_STATUS_RESPONSE  = "Alive and well"
	AUCTION_DURATION         = 3
	SUCCESFUL_MAJORITY       = "succesfully got the bid from more than half of the replicas"
	MAJORITY_WAS_UNSUCCESFUL = "replicas couldn't agree on the current bid"
	SERVER_LOG_DIRECTORY     = "../Server"
)

type AuctionType struct {
	Bid    int32
	Bidder int32
	done   bool
}

type AuctionFromLog struct {
	latestAuction AuctionType
	replicaPort   int32
}

type FrontEndServer struct {
	Auction.UnimplementedAuctionServiceServer
	auctionDone        bool
	replicaServerPorts map[int]bool
}

func main() {
	replicaServerMap := make(map[int]bool)
	feServer := FrontEndServer{auctionDone: false, replicaServerPorts: replicaServerMap}

	go listen(&feServer)
	go feServer.StartAuction()
	for {
		feServer.FindReplicasAndAddThemToMap()
	}
}
func (feServer *FrontEndServer) StartAuction() {
	time.Sleep(time.Minute * AUCTION_DURATION)
	feServer.auctionDone = true
	feServer.EndAuction()
}

func (feServer *FrontEndServer) EndAuction() {
	highestBid, status := feServer.GetHighestBidFromReplicas()
	if strings.Contains(status, "succesfully") {
		lastBid := highestBid
		lastBid.done = true
		fmt.Println(lastBid)
		feServer.UpdateAllReplicas(lastBid)
	} else {
		time.Sleep(time.Second * AUCTION_DURATION)
		fmt.Printf("%v status: %s", highestBid, status)
	}
}
func (feServer *FrontEndServer) Bid(ctx context.Context, message *Auction.BidRequest) (*Auction.BidResponse, error) {
	currentBid, status := feServer.GetHighestBidFromReplicas()
	fmt.Println(status)

	newBid := message.GetAmount()
	if currentBid.Bid < newBid && !feServer.auctionDone {
		feServer.UpdateBid(newBid, message.GetClientId())
		return &Auction.BidResponse{Success: true}, nil
	}
	return &Auction.BidResponse{Success: false}, nil

}

func (feServer *FrontEndServer) Result(ctx context.Context, message *Auction.ResultRequest) (*Auction.ResultResponse, error) {
	currentBid, status := feServer.GetHighestBidFromReplicas()
	err := errors.New("couldn't find any bids on the running auction")
	var response Auction.ResultResponse
	if currentBid.done {
		response = Auction.ResultResponse{BidderID: currentBid.Bidder, HighestBid: currentBid.Bid, Done: true}
	} else if strings.Contains(status, "succesfully") {
		response = Auction.ResultResponse{BidderID: currentBid.Bidder, HighestBid: currentBid.Bid, Done: false}
	} else if strings.Contains(status, "replicas couldnt") {
		return &Auction.ResultResponse{BidderID: -1, HighestBid: -1, Done: false}, err
	}

	return &response, nil

}

func (feServer *FrontEndServer) UpdateBid(bid int32, bidderId int32) {
	newAuction := AuctionType{Bid: bid, Bidder: bidderId, done: false}
	msg := feServer.UpdateAllReplicas(newAuction)
	fmt.Println(msg)
}

func (feServer *FrontEndServer) CreateAuctionFromLog(logMsg string) AuctionFromLog {
	splitMsg := strings.Fields(logMsg)
	var bid, bidder, replicaPort int
	for i, subMsg := range splitMsg {
		switch subMsg {
		case "HighestBid:":
			bid, _ = strconv.Atoi(splitMsg[i+1])
		case "id:":
			bidder, _ = strconv.Atoi(splitMsg[i+1])
		case "port:":
			replicaPort, _ = strconv.Atoi(splitMsg[i+1])
		}
	}
	auction := AuctionType{Bid: int32(bid), Bidder: int32(bidder), done: feServer.auctionDone}
	return AuctionFromLog{latestAuction: auction, replicaPort: int32(replicaPort)}
}

func (feServer *FrontEndServer) GetHighestBidFromReplicas() (AuctionType, string) {
	// get bid from all replicas maybe use timestamps or just r = w = n-1/2
	// the value repeating the most or quorum wins
	latestLogs := feServer.ReadFromLog()
	latestAuctionsFromLogs := make([]AuctionFromLog, 10)
	for _, Msg := range latestLogs {
		if strings.Contains(Msg, "replica port:") {
			fmt.Println(Msg)
			auction := feServer.CreateAuctionFromLog(Msg)
			latestAuctionsFromLogs = append(latestAuctionsFromLogs, auction)
		}
	}

	Quorom := make(map[AuctionType]int)
	for _, auctions := range latestAuctionsFromLogs {
		if auctions.replicaPort >= 5000 {
			fmt.Println(auctions)
			auction := auctions.latestAuction
			temp := Quorom[auction]
			Quorom[auction] = (temp + 1)
		}
	}
	sizeOfMap := 0
	for _, amountVotes := range Quorom {
		sizeOfMap += amountVotes
	}
	for auction, amountVotes := range Quorom {
		if amountVotes > sizeOfMap/2 {
			return auction, SUCCESFUL_MAJORITY
		}

	}
	return AuctionType{Bid: -1, Bidder: -1, done: true}, MAJORITY_WAS_UNSUCCESFUL
}

func (feServer *FrontEndServer) ReadFromLog() []string {
	replicaMsgs := make([]string, len(feServer.replicaServerPorts))
	err := os.Chdir(SERVER_LOG_DIRECTORY)
	if err != nil {
		fmt.Println(err.Error())
	}
	for replica, alive := range feServer.replicaServerPorts {
		if alive {
			t, err := tail.TailFile(fmt.Sprintf("%s%v", SERVER_LOG_FILE, replica), tail.Config{MustExist: true, Follow: true})
			if err != nil {
				return replicaMsgs
			}

			var latestMsg string
			go t.StopAtEOF()

			for line := range t.Lines {
				latestMsg = line.Text
			}
			latestMsgAndPort := fmt.Sprintf("%s replica port: %v", latestMsg, replica)
			replicaMsgs = append(replicaMsgs, latestMsgAndPort)
		}
	}
	return replicaMsgs

}

// update all replica logs
func (feServer *FrontEndServer) UpdateAllReplicas(auction AuctionType) string {
	failedWrites := 0
	ctx := context.Background()
	request := Replica.Auction{Bid: auction.Bid, BidId: auction.Bidder}

	for port, alive := range feServer.replicaServerPorts {
		if alive {
			conn, status := Connect(port)
			if status == "succes" {
				replicaClient, status := ConnectToReplicaClient(conn)
				response, err := replicaClient.WriteToLog(ctx, &request)
				if err != nil {
					fmt.Println(err.Error())
				} else if response.GetMsg() != "ack" || status == "failed" {
					failedWrites++
				}
				conn.Close()
			}
		}
	}
	if failedWrites < len(feServer.replicaServerPorts)/2 {
		return "succesfully updated replica logs"
	}
	return "failed to update more than half of the replicas"
}

// find replicas and add to map
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
			}
		}
		counter++
	}
	ctx.Done()
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
