package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/lottejd/DISYSMP3/Auction"
	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

const (
	clientPort    = 8080
	serverPort    = 5000
	serverLogFile = "serverLog"
	connectionNil = "TRANSIENT_FAILURE"
)

type Server struct {
	Auction.UnimplementedAuctionServiceServer
	Replica.UnimplementedReplicaServiceServer
	id         int32
	primary    bool
	port       int32
	allServers map[int32]Server
	alive      bool
	this       *AuctionType
}

func main() {

	//init
	_, primaryReplicaConn := connect(serverPort)

	fmt.Println(primaryReplicaConn.GetState().String())
	server := CreateReplica(primaryReplicaConn)

	//setup listen on port
	go Listen(server.port, &server)
	go func() {
		for {
			fmt.Println("finding allServers")
			server.findServers()
			time.Sleep(time.Second * 5)
			server.displayAllReplicas()
			time.Sleep(time.Second * 5)
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
	}
	fmt.Scanln()

}

func (s *Server) displayAllReplicas() {
	for _, server := range s.allServers {
		if server.alive {
			fmt.Println(server.ToString())
		}
	}
}

func (s *Server) WriteToLog(ctx context.Context, auction *Replica.Auction) (*Replica.Ack, error) {
	msg := fmt.Sprintf("HighestBid: %v, placed by: %v", auction.Bid, auction.BidId)
	Logger(msg, serverLogFile+strconv.Itoa(int(s.id)))
	return &Replica.Ack{Ack: "ack"}, nil
}

func (s *Server) CheckLeaderStatus(ctx context.Context, request *Replica.GetStatusRequest) (*Replica.LeaderStatusResponse, error) {
	// implement
	return nil, nil
}
func (s *Server) ChooseNewLeader(ctx context.Context, request *Replica.WantToLeadRequest) (*Replica.StatusResponse, error) {
	// implement
	return nil, nil
}

// create a replicaServer
func CreateReplica(conn *grpc.ClientConn) Server {
	var primary bool
	var id, port int32
	var allServers map[int32]Server

	client := Replica.NewReplicaServiceClient(conn)
	replicaInfo, err := client.CreateNewReplica(context.Background(), &Replica.EmptyRequest{})

	if err != nil {
		port = EvalPort(conn)
		primary = evalPrimary(conn)
		allServers = evalServers(conn, replicaInfo)
		id = evalServerId(conn)
	} else {
		id = replicaInfo.GetServerId()
		primary = false
		port = replicaInfo.GetPort()
		allServers = evalServers(conn, replicaInfo)
	}

	auction := AuctionType{0, -1, false} // fjern denne type auction køre gennemm log

	return Server{id: id, primary: primary, port: port, allServers: allServers, alive: true, this: &auction}
}
func (s *Server) ToString() string {
	return fmt.Sprintf("Server id: %v, server port: %v, server status: %v,", s.id, s.port, s.alive)
}

func EvalPort(conn *grpc.ClientConn) int32 {
	var port int32
	if conn.GetState().String() != connectionNil {
		client := Replica.NewReplicaServiceClient(conn)
		response, _ := client.CreateNewReplica(context.Background(), &Replica.EmptyRequest{})
		port = response.GetPort()
	} else {
		port = serverPort
	}
	return port
}

// ask leader for info to create a new replica
func (s *Server) CreateNewReplica(ctx context.Context, emptyRequest *Replica.EmptyRequest) (*Replica.ReplicaInfo, error) {
	var highestId int32
	var highestPort int32
	for _, server := range s.allServers {
		highestId = Max(highestId, server.id)
		highestPort = Max(highestPort, server.port)
	}
	replicas := s.ServerMapToReplicaInfoArray()
	response := Replica.ReplicaInfo{ServerId: (highestId + 1), Port: (highestPort + 1), Replicas: replicas}

	// add new server to map
	temp := CreateTempReplica((highestId + 1), (highestPort + 1))
	s.allServers[highestId] = *temp
	fmt.Println(s.ToString())
	return &response, nil
}

func CreateTempReplica(id int32, port int32) *Server {
	tempReplica := Server{id: id, primary: false, port: port, allServers: nil, alive: true, this: nil}
	return &tempReplica
}

func evalServerId(conn *grpc.ClientConn) int32 {
	if conn.GetState().String() != connectionNil {
		client := Replica.NewReplicaServiceClient(conn)
		response, _ := client.CreateNewReplica(context.Background(), &Replica.EmptyRequest{})
		return response.GetServerId()
	}
	return 0
}

func connect(port int32) (int32, *grpc.ClientConn) {
	// The first attempt will return an error witch will give the first replica ID 0
	// After that it will connect to the port given as a parameter
	conn, err := grpc.Dial(FormatAddress(port), grpc.WithInsecure())
	if err != nil {
		return 0, nil
	}

	client := Replica.NewReplicaServiceClient(conn)

	request := Replica.GetStatusRequest{ServerId: int32(-1)}
	id, err := client.CheckStatus(context.Background(), &request)

	return id.GetServerId(), conn
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
	fmt.Println("hello I'm not blocked")

	// start auction service
	if s.primary {
		grpcServer := grpc.NewServer()
		lis, _ := net.Listen("tcp", FormatAddress(clientPort))
		Auction.RegisterAuctionServiceServer(grpcServer, s)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server over port %v  %v", port, err)
		}

	}
}

func evalPrimary(conn *grpc.ClientConn) bool {
	if conn.GetState().String() != connectionNil {
		return false
	}
	return true
}

func (s *Server) findServers() {
	for id, server := range s.allServers {
		if server.id == s.id {
			continue
		}
		serverId, conn := connect(serverPort)
		if conn.GetState().String() != connectionNil {
			if serverId == id {
				server.alive = true
			}
			continue
		}
		server.alive = false
	}
}

func evalServers(conn *grpc.ClientConn, replicaInfo *Replica.ReplicaInfo) map[int32]Server {
	serverMap := make(map[int32]Server)
	if conn.GetState().String() != connectionNil {
		fmt.Println("hello evalServers")
		for _, replica := range replicaInfo.Replicas {
			serverMap[replica.GetServerId()] = *CreateTempReplica(replica.GetServerId(), replica.GetPort())
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
