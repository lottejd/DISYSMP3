package main

import (
	"context"
	"time"

	"github.com/lottejd/DISYSMP3/Auction"
	"github.com/lottejd/DISYSMP3/Replica"
)

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
	s.UpdateReplicas(bid, bidId)
}

func (s *Server) StartAuction() {
	s.this.done = false
}

func (s *Server) EndAuction() {
	s.this.done = true
}

func (s *Server) UpdateReplicas(bid int32, bidId int32) string {
	var succes bool // fjern efter testing
	if s.primary {
		ctx := context.Background()
		request := Replica.Auction{Bid: bid, BidId: bidId}
		
		for _, server := range s.allServers {
			if server.alive {
				replicaClient, status := ConnectToClient(server.port)
				if status == "failed" {
					server.alive = false
				} else {
					ack, _ := replicaClient.WriteToLog(ctx, &request)
					if ack.GetAck() == "ack" {
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

func ConnectToClient(port int32) (Replica.ReplicaServiceClient, string) {
	var status string
	var client Replica.ReplicaServiceClient
	status = "failed"
	for i := 0; i < 3; i++ {
		_, conn := connect(port)
		if conn == nil {
			time.Sleep(time.Millisecond * 250)
			continue
		}
		client = Replica.NewReplicaServiceClient(conn)
		status = "ack"
		break
	}

	return client, status
}
