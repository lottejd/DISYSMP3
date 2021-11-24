package main

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

func (s *Server) WriteToLog(ctx context.Context, auction *Replica.Auction) (*Replica.AuctionAck, error) {
	msg := fmt.Sprintf("HighestBid: %v, placed by: %v", auction.Bid, auction.BidId)
	Logger(msg, ServerLogFile+strconv.Itoa(int(s.id)))

	return &Replica.AuctionAck{Bid: auction, Msg: "ack"}, nil
}

func IsConnectable(conn *grpc.ClientConn) bool {
	return conn.GetState().String() != ConnectionNil
}

func StartReplicaService(port int32, s *Server) {
	lis, _ := net.Listen("tcp", FormatAddress(s.port))
	defer lis.Close()

	grpcServer := grpc.NewServer()
	Replica.RegisterReplicaServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		fmt.Printf("Failed to serve %v on port %v\n", s.id, s.port)
	}
}
