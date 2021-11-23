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

//  this is a little iffy until we ensure new replicas take over port 5000
func EvalPrimary(conn *grpc.ClientConn) bool {
	return !IsConnectable(conn)
}

// maybe this is redundant since FindServersAndAddToMap() does this anyways
func EvalServers(conn *grpc.ClientConn, replicaInfo *Replica.ReplicaInfo) map[int32]Server {
	serverMap := make(map[int32]Server)
	if IsConnectable(conn) {
		for _, replica := range replicaInfo.Replicas {
			serverMap[replica.GetServerId()] = *CreateProxyReplica(replica.GetServerId(), replica.GetPort())
		}
	}
	return serverMap
}

// maybe this is redundant since FindServersAndAddToMap() does this anyways
func (s *Server) ServerMapToReplicaInfoArray() []*Replica.ReplicaInfo {
	var servers []*Replica.ReplicaInfo
	for _, Server := range s.allServers {
		temp := Replica.ReplicaInfo{ServerId: Server.id, Port: Server.port}
		servers = append(servers, &temp)
	}
	return servers
}

func (s *Server) FindServersAndAddToMap() {
	for i := 0; i < 10; i++ {
		if int(s.id) == i {
			// if s.primary {
			// 	fmt.Println("")
			// 	// remove and add replica to update the map 
			// 	temp := s.allServers[s.id]
			// 	temp.SetPrimary()
			// 	s.RemoveReplicaFromMap(s.id)
			// 	s.AddReplicaToMap(&temp)
			// }
			continue
		}

		serverId, conn := Connect(int32(ServerPort + i))
		if IsConnectable(conn) {
			replica := CreateProxyReplica(serverId, int32(ServerPort+i))
			replicaClient, _ := ConnectToReplicaClient(replica.port)

			request := Replica.GetStatusRequest{ServerId: replica.id}
			response, _ := replicaClient.CheckStatus(context.Background(), &request)
			if response.Primary {
				replica.SetPrimary()
			}
			s.AddReplicaToMap(replica)
		} else {
			// if no connection is 
			s.RemoveReplicaFromMap(serverId)
		}
	}
}

func (s *Server) RemoveReplicaFromMap(serverId int32) {
	delete(s.allServers, serverId)
}

// split this method into StartReplicaService and StartAuctionService
// and add retrys for StartReplicaService
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

	grpcServer := grpc.NewServer()
	Replica.RegisterReplicaServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve on")
	}

}

func StartAuctionService(s *Server) {
	grpcServer := grpc.NewServer()
	lis, _ := net.Listen("tcp", FormatAddress(ClientPort))
	defer lis.Close()

	Auction.RegisterAuctionServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port %v  %v", ClientPort, err)
	}
}
