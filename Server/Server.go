package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/lottejd/DISYSMP3/Auction"
	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

type Server struct {
	Replica.UnimplementedReplicaServiceServer
	Auction.UnimplementedAuctionServiceServer
	id         int32
	primary    bool
	port       int32
	allServers map[int32]Server
	alive      bool
	this       *AuctionType
}

func main() {

	//init
	_, primaryReplicaConn := Connect(ServerPort)
	server := CreateReplica(primaryReplicaConn)
	ctx := context.Background()

	//setup listen on port
	go Listen(server.port, &server)
	go func() {
		for {
			go server.DisplayAllReplicas()
			time.Sleep(time.Second * 5)
			go server.FindServers()

		}
	}()

	// start the service / server on the specific port
	// skal rykkes ud i en go routine, i tilfæde primary dør bliver det her ikke kørt igen
	fmt.Println(server.primary)
	if server.primary {
		server.allServers[server.id] = server
		for {
			server.StartAuction()
			time.Sleep(time.Minute * 5)
			server.EndAuction()
		}
	} else {
		for _, s := range server.allServers {
			if server.id != s.id {
				client, _ := ConnectToReplicaClient(s.port)
				response, _ := client.CheckStatus(ctx, &Replica.GetStatusRequest{ServerId: server.id})
				if response.GetPrimary() {
					temp := s
					temp.SetPrimary()
					server.allServers[s.id] = temp
				}
			}

		}
	}
	fmt.Scanln()
}

func (s *Server) CheckStatus(ctx context.Context, request *Replica.GetStatusRequest) (*Replica.StatusResponse, error) {
	// implement
	replicas := s.ServerMapToReplicaInfoArray()
	response := Replica.StatusResponse{ServerId: s.id, Primary: s.primary, Replicas: replicas}
	return &response, nil
}
func (s *Server) ChooseNewLeader(ctx context.Context, request *Replica.WantToLeadRequest) (*Replica.StatusResponse, error) {
	// implement
	return nil, nil
}

// create a replicaServer
func CreateReplica(conn *grpc.ClientConn) Server {

	var id, port int32
	client := Replica.NewReplicaServiceClient(conn)
	replicaInfo, err := client.CreateNewReplica(context.Background(), &Replica.EmptyRequest{})

	primary := EvalPrimary(conn)
	allServers := EvalServers(conn, replicaInfo)
	if err != nil {
		fmt.Println(err.Error())
		port = ServerPort
		id = 0
	} else {
		id = replicaInfo.GetServerId()
		port = replicaInfo.GetPort()
	}

	auction := AuctionType{0, -1, false} // fjern denne type auction køre gennemm log
	s := Server{id: id, primary: primary, port: port, allServers: allServers, alive: true, this: &auction}
	s.AddReplica(&s)
	s.DisplayAllReplicas()
	return s
}

// ask leader for info to create a new replica
func (s *Server) CreateNewReplica(ctx context.Context, emptyRequest *Replica.EmptyRequest) (*Replica.ReplicaInfo, error) {
	var highestId int32
	var highestPort int32
	for _, server := range s.allServers {
		highestId = Max(highestId, server.id)
		highestPort = Max(highestPort, server.port)
	}
	highestId += 1
	highestPort += 1
	replicas := s.ServerMapToReplicaInfoArray()
	response := Replica.ReplicaInfo{ServerId: (highestId), Port: (highestPort), Replicas: replicas}

	// add new server to map
	temp := CreateProxyReplica((highestId), (highestPort))
	s.allServers[highestId] = *temp
	return &response, nil
}

func Listen(port int32, s *Server) {
	// start peer to peer service
	go func() {
		lis, _ := net.Listen("tcp", FormatAddress(port))

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
		Auction.RegisterAuctionServiceServer(grpcServer, s)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server over port %v  %v", port, err)
		}

	}
}
