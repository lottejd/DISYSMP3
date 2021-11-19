package main

import (
	"context"
	"fmt"
	"log"
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

func EvalPrimary(conn *grpc.ClientConn) bool {
	return !IsConnectable(conn)
}

func EvalServers(conn *grpc.ClientConn, replicaInfo *Replica.ReplicaInfo) map[int32]Server {
	serverMap := make(map[int32]Server)
	if IsConnectable(conn) {
		for _, replica := range replicaInfo.Replicas {
			serverMap[replica.GetServerId()] = *CreateProxyReplica(replica.GetServerId(), replica.GetPort())
		}
	}
	return serverMap
}

func (s *Server) ServerMapToReplicaInfoArray() []*Replica.ReplicaInfo {
	var servers []*Replica.ReplicaInfo
	for _, Server := range s.allServers {
		temp := Replica.ReplicaInfo{ServerId: Server.id, Port: Server.port}
		servers = append(servers, &temp)
	}
	return servers
}

func (s *Server) FindServers() {
	for i := 0; i < 10; i++ {
		if int(s.id) == i {
			continue
		}
		serverId, conn := Connect(int32(ServerPort + i))
		if IsConnectable(conn) && s.allServers[int32(serverId)].port == 0 {
			replica := CreateProxyReplica(serverId, int32(ServerPort+i))
			s.AddReplica(replica)
		}
	}
}

func Listen(port int32, s *Server) {
	// start peer to peer service
	go func() {
		lis, _ := net.Listen("tcp", FormatAddress(port))
		defer lis.Close()

		grpcServer := grpc.NewServer()
		Replica.RegisterReplicaServiceServer(grpcServer, s)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve on")
		}
	}()

	// start auction service
	if s.primary {
		grpcServer := grpc.NewServer()
		lis, _ := net.Listen("tcp", FormatAddress(ClientPort))
		defer lis.Close()

		Auction.RegisterAuctionServiceServer(grpcServer, s)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server over port %v  %v", port, err)
		}

	}
}
