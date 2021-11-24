package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hpcloud/tail"
	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

type Server struct {
	Replica.UnimplementedReplicaServiceServer
	id             int32
	primary        bool
	port           int32
	allServers     map[int32]Server
	alive          bool
	isPrimaryDead  chan bool
	replicaService Service
}

func main() {

	//init
	_, primaryReplicaConn := Connect(ServerPort)
	server := CreateReplica(primaryReplicaConn)
	server.isPrimaryDead = make(chan bool)
	ctx := context.Background()

	//setup listen on port
	fmt.Println(server.port)
	go StartReplicaService(server.port, &server)

	// find and display all replicas
	go Print(&server)

	if server.primary {
		server.allServers[server.id] = server
		go server.PrimaryReplicaLoop()
	} else {
		go server.ReplicaLoop()

		for {
			fmt.Println("hello 1")
			temp := <-server.isPrimaryDead
			fmt.Printf("Server chan says isPrimaryDead: %v\n", temp)
			if temp {
				fmt.Println("leader is dead find a new one")
				server.KillLeaderLocally()
				outcome := server.StartElection()
				if outcome == "Winner" {
					fmt.Println(outcome)
					server.SetPrimary()
					server.port = ServerPort
					go StartReplicaService(ServerPort, &server)
					go server.PrimaryReplicaLoop()
					break
				}
			}
		}
	}

	for server.primary {
		fmt.Println("inside last loop")
		temp := <-server.isPrimaryDead
		fmt.Println("bye main")
		server.primary = !temp
		CloseConnectionAndCtx(ctx, primaryReplicaConn)
	}

}

func (s *Server) PrimaryReplicaLoop() {
	var KillThisServer bool
	fmt.Printf("New primary id: %v\n", s.id)
	t, err := tail.TailFile(ServerLogFile+strconv.Itoa(int(s.id)), tail.Config{Follow: true})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	var latestBid string
	fmt.Println(t.Filename)
	fmt.Println(latestBid)
	for line := range t.Lines {
		latestBid = line.Text
	}
	fmt.Println(latestBid)
	if latestBid != "" {
		split := strings.Fields(latestBid)
		bid, _ := strconv.Atoi(split[3])
		bidder, _ := strconv.Atoi(split[6])
		fmt.Println(split[3])
		s.this.highestBid = int32(bid)
		s.this.highestBidder = int32(bidder)
	} else {
		//gør noget med seneste bud
		go func() {
			fmt.Println("starting auction")
			s.StartAuction()
			time.Sleep(time.Minute * 3)
			fmt.Println("Ending auction")
			s.EndAuction()
		}()
	}
	for !KillThisServer {
		temp := <-s.isPrimaryDead
		println("Primary should die")
		KillThisServer = temp
	}
}

func (s *Server) ReplicaLoop() {
	go KillPrimaryFromClient(s)
	for {
		if s.primary {
			fmt.Println("im breaking the replica loop to become primary")
			s.replicaService.stopService <- true
			break
		}

		leaderIsDead := true
		ctx := context.Background()
		// Check if leader is dead
		for _, server := range s.allServers {
			if s.id != server.id {
				client, _ := ConnectToReplicaClient(server.port)
				response, _ := client.CheckStatus(ctx, &Replica.GetStatusRequest{ServerId: s.id})
				if response.GetPrimary() {
					leaderIsDead = false
					temp := server
					temp.SetPrimary()
					s.allServers[server.id] = temp
				}
			} else if len(s.allServers) == 1 {
				leaderIsDead = false
			}
			ctx.Done()
		}
		if leaderIsDead {
			s.isPrimaryDead <- leaderIsDead // use a channel to break the loop in main
		}
		time.Sleep(time.Second * 2)
	}
}

// gRPC service sends replicas own status
func (s *Server) CheckStatus(ctx context.Context, request *Replica.GetStatusRequest) (*Replica.StatusResponse, error) {
	response := Replica.StatusResponse{ServerId: s.id, Primary: s.primary}
	return &response, nil
}

func (s *Server) KillPrimary(ctx context.Context, empty *Replica.EmptyRequest) (*Replica.Answer, error) {
	if s.primary {
		fmt.Printf("I am getting killed %v\n", s.id)
		s.isPrimaryDead <- true
		return &Replica.Answer{ServerId: s.id, Alive: false}, nil
	}
	return &Replica.Answer{ServerId: s.id, Alive: s.alive}, nil
}

// begin Election bully style
func (s *Server) StartElection() string {
	msg := Replica.ElectionMessage{ServerId: s.id, Msg: "Election"}
	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelCtx()
	for _, server := range s.allServers {
		if s.id < server.id {
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
	return &response, nil
}

// create a replicaServer
func CreateReplica(conn *grpc.ClientConn) Server {
	var id, port int32
	client := Replica.NewReplicaServiceClient(conn)
	ctx := context.Background()
	replicaInfo, err := client.CreateNewReplica(ctx, &Replica.EmptyRequest{})

	primary := EvalPrimary(conn)
	allServers := make(map[int32]Server)
	fmt.Println(IsConnectable(conn))
	if !IsConnectable(conn) {
		fmt.Println(err)
		port = ServerPort
		id = 0
	} else {
		id = replicaInfo.GetServerId()
		port = replicaInfo.GetPort()
	}

	CloseConnectionAndCtx(ctx, conn)                                                       // fjern denne type auction køre gennemm log
	s := Server{id: id, primary: primary, port: port, allServers: allServers, alive: true} //arbiter: sync.Mutex{}
	s.AddReplicaToMap(&s)
	return s
}

// ask leader for info to create a new replica (gRPC)
func (s *Server) CreateNewReplica(ctx context.Context, emptyRequest *Replica.EmptyRequest) (*Replica.ReplicaInfo, error) {
	var highestId int32
	var highestPort int32
	for _, server := range s.allServers {
		highestId = Max(highestId, server.id)
		highestPort = Max(highestPort, server.port)
	}
	highestId += 1
	highestPort += 1
	response := Replica.ReplicaInfo{ServerId: (highestId), Port: (highestPort)}

	// add new server to map
	temp := CreateProxyReplica((highestId), (highestPort))
	s.allServers[highestId] = *temp
	return &response, nil
}
