package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
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
	arbiter    sync.Mutex
	this       *AuctionType
}

func main() {

	//init
	leaderIsDead := make(chan bool)
	_, primaryReplicaConn := Connect(ServerPort)
	server := CreateReplica(primaryReplicaConn)

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
	if server.primary {
		server.allServers[server.id] = server
		for {
			server.StartAuction()
			time.Sleep(time.Minute * 5)
			server.EndAuction()
		}
	} else {
		go server.ReplicaLoop(leaderIsDead)

		for {
			temp := <-leaderIsDead
			if temp {
				fmt.Println("leader is dead find a new one")
				server.KillLeader()
				outcome := server.StartElection()
				fmt.Println(outcome)
				if outcome == "Winner" {
					UpdatePrimaryReplica(&server)
					server.PrimaryLoop()
					break
				}
			}
		}

	}

	fmt.Scanln()
}

func UpdatePrimaryReplica(s *Server) {
	s.arbiter.Lock()
	s.KillLeader()
	s.SetPrimary()
	s.arbiter.Unlock()
	go Listen(ServerPort, s)
}

func (s *Server) PrimaryLoop() {
	fmt.Println("hello loooop")
	for {
		time.Sleep(time.Second * 2)
		fmt.Println(s.ToString())
	}
}

func (s *Server) ReplicaLoop(leaderStatus chan bool) {
	for {
		if s.primary {
			UpdatePrimaryReplica(s)
			break
		}
		leaderIsDead := true
		ctx := context.Background()
		for _, server := range s.allServers {
			if s.id != server.id && server.alive {
				client, _ := ConnectToReplicaClient(server.port)
				response, _ := client.CheckStatus(ctx, &Replica.GetStatusRequest{ServerId: s.id})
				if response.GetPrimary() {
					leaderIsDead = false
					temp := server
					temp.SetPrimary()
					s.allServers[server.id] = temp
				}
			}
		}
		time.Sleep(time.Millisecond * 500)
		if leaderIsDead {
			leaderStatus <- leaderIsDead
		}
		time.Sleep(time.Second * 2)
	}
}

func (s *Server) CheckStatus(ctx context.Context, request *Replica.GetStatusRequest) (*Replica.StatusResponse, error) {
	replicas := s.ServerMapToReplicaInfoArray()
	response := Replica.StatusResponse{ServerId: s.id, Primary: s.primary, Replicas: replicas}
	return &response, nil
}
func (s *Server) ChooseNewLeader(ctx context.Context, request *Replica.WantToLeadRequest) (*Replica.VoteResponse, error) {
	// implement
	return nil, nil
}

func (s *Server) StartElection() string {
	msg := Replica.ElectionMessage{ServerId: s.id, Msg: "Election"}
	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelCtx()
	for _, server := range s.allServers {
		fmt.Println("server range loop")
		if s.id > server.id {
			client, ack := ConnectToReplicaClient(server.port)
			fmt.Print(ack)
			response, err := client.Election(ctx, &msg)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			if response.Alive {
				return "Lost"
			}
		}
	}
	return "Winner"
}

func (s *Server) Election(ctx context.Context, msg *Replica.ElectionMessage) (*Replica.Answer, error) {
	// implement
	electionId := msg.GetServerId()
	if strings.EqualFold("Winner", msg.GetMsg()) && electionId > s.id {
		temp := s.allServers[electionId]
		temp.SetPrimary()
		s.allServers[electionId] = temp
	}
	response := Replica.Answer{ServerId: s.id, Alive: s.alive}

	if electionId < s.id {
		outcome := s.StartElection()
		if strings.EqualFold(outcome, "Winner") {
			msg := Replica.ElectionMessage{ServerId: s.id, Msg: "Winner"}
			ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*10)
			defer cancelCtx()
			for _, server := range s.allServers {
				client, _ := ConnectToReplicaClient(server.port)
				response, _ := client.Election(ctx, &msg)
				if response.GetServerId() < s.id {
					break
				}
			}
			s.SetPrimary()
		}
	}

	// s is small guy just wait no response :-(
	fmt.Println(response.GetServerId())
	return &response, nil
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
	s := Server{id: id, primary: primary, port: port, allServers: allServers, alive: true, arbiter: sync.Mutex{}, this: &auction}
	s.AddReplica(&s)
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
