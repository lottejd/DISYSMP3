package main

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/lottejd/DISYSMP3/Auction"
	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

func (s *Server) WriteToLog(ctx context.Context, auction *Replica.Auction) (*Replica.Ack, error) {
	msg := fmt.Sprintf("HighestBid: %v, placed by: %v", auction.Bid, auction.BidId)
	Logger(msg, ServerLogFile+strconv.Itoa(int(s.id)))
	return &Replica.Ack{Ack: "ack"}, nil
}

func IsConnectable(conn *grpc.ClientConn) bool {
	return conn.GetState().String() != ConnectionNil
}

//  this is a little iffy until we ensure new replicas take over port 5000
func EvalPrimary(conn *grpc.ClientConn) bool {
	return !IsConnectable(conn)
}
func (s *Server) FindServersAndAddToMap() {
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		if int(s.id) == i {
			s.allServers[s.id] = *s
		} else {
			serverId, conn := Connect(int32(ServerPort + i))
			if IsConnectable(conn) {
				replica := CreateProxyReplica(serverId, int32(ServerPort+i))
				replicaClient, _ := ConnectToReplicaClient(replica.port)

				request := Replica.GetStatusRequest{ServerId: replica.id}
				response, _ := replicaClient.CheckStatus(ctx, &request)
				if response.Primary {
					replica.SetPrimary()
				}
				s.AddReplicaToMap(replica)
				conn.Close()
			} else {
				// fmt.Printf("removing: %v from the map\n", i)
				s.RemoveReplicaFromMap(int32(i))
			}
		}
	}
	ctx.Done()
}

func (s *Server) RemoveReplicaFromMap(serverId int32) {
	delete(s.allServers, serverId)
}

func Listen(port int32, s *Server) {
	// start peer to peer service
	go StartReplicaService(port, s)

	// start auction service
	if s.primary {
		StartAuctionService(s)

	}
}

func StartReplicaService(port int32, s *Server) {
	lis, _ := net.Listen("tcp", FormatAddress(port))
	defer lis.Close()

	grpcServer := grpc.NewServer()
	Replica.RegisterReplicaServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		fmt.Printf("Failed to serve %v on port %v\n", s.id, port)
	}

}

func StartAuctionService(s *Server) {
	grpcServer := grpc.NewServer()
	lis, _ := net.Listen("tcp", FormatAddress(ClientPort))
	defer lis.Close()

	Auction.RegisterAuctionServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		fmt.Printf("Failed to serve %v on port %v\n", s.id, ClientPort)
	}
}
