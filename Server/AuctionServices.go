package main

import (
	"context"
	"strings"
	"time"

	"github.com/lottejd/DISYSMP3/Auction"
	"github.com/lottejd/DISYSMP3/Replica"
)

//server side
func (s *Server) Bid(ctx context.Context, message *Auction.BidRequest) (*Auction.BidResponse, error) {
	if message.Amount > s.this.highestBid {
		s.UpdateBid(message.Amount, message.ClientId)
		return &Auction.BidResponse{Success: true}, nil
	}
	return &Auction.BidResponse{Success: true}, nil
}

func (s *Server) Result(ctx context.Context, message *Auction.ResultRequest) (*Auction.ResultResponse, error) {
	return &Auction.ResultResponse{BidderID: s.this.highestBidder, HighestBid: s.this.highestBid, Done: s.this.done}, nil
}


func (s *Server) UpdateBid(bid int32, bidId int32) {
	s.this.highestBid = bid
	s.this.highestBidder = bidId
	s.UpdateAllReplicaLogs(bid, bidId)
}

func (s *Server) StartAuction() {
	s.this.done = false
}

func (s *Server) EndAuction() {
	s.this.done = true
}

func (s *Server) UpdateAllReplicaLogs(bid int32, bidId int32) string {
	var succes bool
	if s.primary {
		ctx := context.Background()
		request := Replica.Auction{Bid: bid, BidId: bidId}

		for _, server := range s.allServers {
			if server.alive {
				replicaClient, status := ConnectToReplicaClient(server.port)
				if strings.EqualFold(status, "failed") {
					server.alive = false
				} else {
					ack, _ := replicaClient.WriteToLog(ctx, &request)
					if strings.EqualFold(ack.GetAck(), "succes") {
						succes = true
					}
				}
			}
		}
	}
	if succes {
		return "succes"
	}
	return "failed"
}

// connect to a replica with at max 3 retries
func ConnectToReplicaClient(port int32) (Replica.ReplicaServiceClient, string) {
	var status string
	var client Replica.ReplicaServiceClient
	status = "failed"
	for i := 0; i < 3; i++ {
		_, conn := Connect(port)
		if conn == nil {
			time.Sleep(time.Millisecond * 250)
			continue
		}
		client = Replica.NewReplicaServiceClient(conn)
		status = "succes"
		break
	}

	return client, status
}
