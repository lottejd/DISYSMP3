package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/lottejd/DISYSMP3/Auction"
	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

const ClientPort = 8080

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
	listen(&feServer)

	fmt.Scanln()
	fmt.Println("Main FrontEnd is now Done")
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
	feServer.UpdateAllReplicas(newAuction) // implement
	feServer.this = &newAuction
}

func (feServer *FrontEndServer) GetHighestBidFromReplicas() (AuctionType, string) {
	// implement
	// get bid from all replicas maybe use timestamps or just r = w = n-1/2
	// the value repeating the most or quorum wins
	return *feServer.this, "succesfully got the bid from frontend"
}

func (feServer *FrontEndServer) UpdateAllReplicas(auction AuctionType) string {
	// implement
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
				} else if response.GetAck() != "succes" || status == "failed" {
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

func (feServer *FrontEndServer) EvalAuctionStillGoing(current time.Time) bool {
	fmt.Println(feServer.startTime)
	return true
}

// grcp server setup
func listen(feServer *FrontEndServer) {
	address := fmt.Sprintf("localhost:%v", ClientPort)
	lis, _ := net.Listen("tcp", address)
	defer lis.Close()

	grcpServer := grpc.NewServer()
	Auction.RegisterAuctionServiceServer(grcpServer, feServer)
	if err := grcpServer.Serve(lis); err != nil {
		fmt.Printf("Failed to serve AuctionService on port %v\n", ClientPort)
	}
}

// connect to a replica with at max 3 retries
func ConnectToReplicaClient(conn *grpc.ClientConn) (Replica.ReplicaServiceClient, string) {
	ctx := context.Background()
	var client Replica.ReplicaServiceClient
	request := Replica.GetStatusRequest{ServerId: int32(-1)}
	client = Replica.NewReplicaServiceClient(conn)
	ack, err := client.CheckStatus(ctx, &request)
	if err != nil || ack.GetServerId() < 0 {
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
