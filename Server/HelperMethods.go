package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

const (
	ClientPort    = 8080
	ServerPort    = 5000
	ServerLogFile = "serverLog"
	ConnectionNil = "TRANSIENT_FAILURE"
)

type AuctionType struct {
	highestBid    int32
	highestBidder int32
	done          bool
}

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

func Min(this int32, that int32) int32 {
	if this < that {
		return this
	}
	return that
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

func (s *Server) AddReplica(replica *Server) {
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

func (s *Server) KillLeader() {
	for _, server := range s.allServers {
		if server.primary {
			temp := server
			temp.alive = false
			temp.primary = false
			s.allServers[server.id] = temp
			break
		}
	}
}

// deprecated

// func EvalServerId(conn *grpc.ClientConn) int32 {

// 	if IsConnectable(conn) {
// 		client := Replica.NewReplicaServiceClient(conn)
// 		response, _ := client.CreateNewReplica(context.Background(), &Replica.EmptyRequest{})
// 		return response.GetServerId()
// 	}
// 	return 0
// }

// func EvalPort(conn *grpc.ClientConn) int32 {
// 	var port int32
// 	if IsConnectable(conn) {
// 		client := Replica.NewReplicaServiceClient(conn)
// 		response, _ := client.CreateNewReplica(context.Background(), &Replica.EmptyRequest{})
// 		port = response.GetPort()
// 	} else {
// 		port = ServerPort
// 	}
// 	return port
// }
