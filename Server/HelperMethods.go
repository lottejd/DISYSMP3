package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

const (
	ClientPort    = 8080
	ServerPort    = 5000
	ServerLogFile = "serverLog"
	ConnectionNil = "TRANSIENT_FAILURE" // instead of nil when trying to connect to a port without a ReplicaService registered
)

type AuctionType struct {
	highestBid    int32
	highestBidder int32
	done          bool
}

// move this into ConnectToReplicaClient or refactor
func Connect(port int32) (int32, *grpc.ClientConn) {
	// The first attempt will return an error witch will give the first replica ID 0
	// After that it will connect to the port given as a parameter
	conn, err := grpc.Dial(FormatAddress(port), grpc.WithInsecure())
	if err != nil {
		return 0, nil
	}

	client := Replica.NewReplicaServiceClient(conn)

	request := Replica.GetStatusRequest{ServerId: -1}
	id, _ := client.CheckStatus(context.Background(), &request)
	return id.GetServerId(), conn
}

func Logger(message string, logFileName string) {
	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(message)
}

func Max(this int32, that int32) int32 {
	if this < that {
		return that
	}
	return this
}

func FormatAddress(port int32) string {
	address := fmt.Sprintf("localhost:%v", port)
	return address
}

func (s *Server) ToString() string {
	return fmt.Sprintf("Server id: %v, server port: %v, server is primary: %v,  server alive: %v,", s.id, s.port, s.primary, s.alive)
}

func CreateProxyReplica(id int32, port int32) *Server {
	tempReplica := Server{id: id, primary: false, port: port, allServers: nil, alive: true, this: nil}
	return &tempReplica
}

func (s *Server) AddReplicaToMap(replica *Server) {
	s.allServers[replica.id] = *replica
}

func (s *Server) SetPrimary() {
	s.primary = true
}

func (s *Server) DisplayAllReplicas() {
	fmt.Println("Display All Replicas")
	for _, server := range s.allServers {
		fmt.Println(server.ToString())
	}
}

// Mark primary as dead in the replica server map
func (s *Server) KillLeaderLocally() {
	for _, server := range s.allServers {
		if server.primary && server.id != s.id {
			fmt.Println("Did we reach KillLeaderLocally")
			temp := server
			temp.alive = false
			temp.primary = false
			s.allServers[server.id] = temp
		}
	}
}

func KillPrimaryFromClient(s *Server) {
	fmt.Println("Sleeping")
	time.Sleep(time.Second * 30)
	fmt.Println("Done sleeping")
	ctx := context.Background()
	for _, server := range s.allServers {
		ReplicaClient, _ := ConnectToReplicaClient(server.port)
		ReplicaClient.KillPrimary(ctx, &Replica.EmptyRequest{})
	}

}

func waitForInput(s *Server) string {
	fmt.Printf("Replica ID %v - ", s.id)
	// to ensure "enter" has been hit before publishing - skud ud til Mie
	reader, err := bufio.NewReader(os.Stdin).ReadString('\n')
	// remove newline windows format "\r\n"
	if err != nil {
		return "bad input"
	}
	input := strings.TrimSuffix(reader, "\r\n")
	return input
}

func Print(server *Server) {
	for {
		server.DisplayAllReplicas()
		time.Sleep(time.Second * 5)
		server.FindServersAndAddToMap()
	}
}
