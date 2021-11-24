package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"

	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

const (
	SERVER_PORT     = 5000
	SERVER_LOG_FILE = "serverLog"
	// instead of nil when trying to connect to a port without a ReplicaService registered
	CONNECTION_NIL = "TRANSIENT_FAILURE"
)

type Server struct {
	Replica.UnimplementedReplicaServiceServer
	id   int32
	port int32
}

func main() {

	//init
	freeServerPort := FindFreePort()
	fmt.Printf("The next free server port is: %v\n", freeServerPort)
	server := CreateReplica(freeServerPort)

	//setup listen on port
	StartReplicaService(server.port, &server)

	// block
	fmt.Scanln()
	fmt.Println("I'm out of main cya")
}

func CreateReplica(port int32) Server {
	id := rand.Intn(10000)
	server := Server{id: int32(id), port: port}
	return server
}

func FindFreePort() int32 {
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		conn, status := Connect(SERVER_PORT + i)
		if status == "succes" {
			_, ack := ConnectToReplicaClient(conn)
			if ack == "succes" {
				continue
			} else {
				ctx.Done()
				return int32(SERVER_PORT + i)
			}
		} else if status == "failed" {
			ctx.Done()
			return int32(SERVER_PORT + i)
		}
	}
	ctx.Done()
	return -1
}

// gRPC services
func (s *Server) CheckStatus(ctx context.Context, empty *Replica.EmptyRequest) (*Replica.StatusResponse, error) {
	response := Replica.StatusResponse{ServerId: s.id, Msg: "Alive and well"}
	return &response, nil
}

// gRPC services
func (s *Server) WriteToLog(ctx context.Context, auction *Replica.Auction) (*Replica.AuctionAck, error) {
	msg := fmt.Sprintf("HighestBid: %v, placed by: %v", auction.Bid, auction.BidId)
	Logger(msg, fmt.Sprintf("%s%v", SERVER_LOG_FILE, s.port))
	fmt.Println(msg)
	return &Replica.AuctionAck{Bid: auction, Msg: "ack"}, nil
}

func IsConnectable(conn *grpc.ClientConn) bool {
	return conn.GetState().String() != CONNECTION_NIL
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
