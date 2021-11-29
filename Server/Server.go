package main

import (
	"context"
	"fmt"
	"net"

	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

const (
	SERVER_PORT     = 5000
	SERVER_LOG_FILE = "serverLog"
	MAX_REPLICAS_ALLOWED = 10
)

type Server struct {
	Replica.UnimplementedReplicaServiceServer
	port int32
}

func main() {

	//init
	freeServerPort := FindFreePort()
	fmt.Printf("The next free server port is: %v\n", freeServerPort)
	server := Server{port: freeServerPort}

	//setup listen on port
	StartReplicaService(server.port, &server)

	// block
	fmt.Scanln()
	fmt.Println("I'm out of main cya")
}

func FindFreePort() int32 {
	ctx := context.Background()
	for i := 0; i < MAX_REPLICAS_ALLOWED; i++ {
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
	response := Replica.StatusResponse{ServerId: s.port, Msg: "Alive and well"}
	return &response, nil
}

// gRPC services
func (s *Server) WriteToLog(ctx context.Context, auction *Replica.Auction) (*Replica.AuctionAck, error) {
	msg := fmt.Sprintf("HighestBid: %v placed by id: %v", auction.Bid, auction.BidId)
	Logger(msg, fmt.Sprintf("%s%v", SERVER_LOG_FILE, s.port))
	fmt.Println(msg)
	return &Replica.AuctionAck{Bid: auction, Msg: "ack"}, nil
}

func StartReplicaService(port int32, s *Server) {
	lis, _ := net.Listen("tcp", FormatAddress(s.port))
	defer lis.Close()

	grpcServer := grpc.NewServer()
	Replica.RegisterReplicaServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		fmt.Printf("Failed to serve on port %v\n", s.port)
	}
}
