package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	auction "github.com/lottejd/DISYSMP3/Auction"
	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

const (
	clientPort    = 8080
	serverPort    = 5000
	serverLogFile = "serverLog"
)

type Server struct {
	auction.UnimplementedAuctionServiceServer
	Replica.UnimplementedReplicaServiceServer
	id         int
	primary    bool
	port       int
	allServers map[int32]*Server
	this       *Auction
}

type Auction struct {
	highestBid    int32
	highestBidder int32
	done          bool
}

func evalPort() int {
	return serverPort
}

func evalPrimary() bool {
	return true
}

func evalServers() map[int32]*Server {
	return nil
}

func main() {

	//init
	var primary bool
	var port int
	var allServers map[int32]*Server
	auction := Auction{0, -1, false}

	port = evalPort()
	primary = evalPrimary()
	allServers = evalServers()
	id := evalServerId()

	s := Server{id: id, primary: primary, port: port, allServers: allServers, this: &auction}

	//setup listen on port
	go Listen(s.port, &s)

	// start the service / server on the specific port
	if primary {
		for {
			s.StartAuction()
			time.Sleep(time.Minute * 5)
			s.EndAuction()
		}
	}

}

func evalServerId() int {
	id := connect(serverPort)
	return id
}

func connect(port int) int {
	// The first attempt will return an error witch will give the first replica ID 0
	// After that it will connect to the port given as a parameter
	conn, err := grpc.Dial(formatAddress(port), grpc.WithInsecure())
	if err != nil {
		return 0
	}

	ctx := context.Background()
	connect := Replica.NewReplicaServiceClient(conn)

	request := Replica.GetStatusRequest{ServerId: int32(-1)}
	id, err := connect.CheckStatus(ctx, &request)

	return int(id.GetServerId())
}

func Listen(port int, s *Server) {

	go func() {
		lis, _ := net.Listen("tcp", formatAddress(port))

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
		lis, _ := net.Listen("tcp", formatAddress(clientPort))
		auction.RegisterAuctionServiceServer(grpcServer, s)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server over port %v  %v", port, err)
		}

	}
}
func (s *Server) CheckLeaderStatus(ctx context.Context, request *Replica.GetStatusRequest) (*Replica.LeaderStatusResponse, error) {
	// implement
	return nil, nil
}
func (s *Server) ChooseNewLeader(ctx context.Context, request *Replica.WantToLeadRequest) (*Replica.StatusResponse, error) {
	// implement
	return nil, nil
}

func (s *Server) Bid(ctx context.Context, message *auction.BidRequest) (*auction.BidResponse, error) {
	if message.Amount > s.this.highestBid {
		s.updateBid(message.Amount, message.ClientId)
		return &auction.BidResponse{Success: true}, nil
	}
	return &auction.BidResponse{Success: true}, nil
}

func (s *Server) Result(ctx context.Context, message *auction.ResultRequest) (*auction.ResultResponse, error) {
	return &auction.ResultResponse{BidderID: s.this.highestBidder, HighestBid: s.this.highestBid, Done: s.this.done}, nil
}

func (s *Server) updateBid(bid int32, bidid int32) {
	s.this.highestBid = bid
	s.this.highestBidder = bidid
	//reportToPrimary()
}

func (s *Server) StartAuction() {
	s.this.done = false
}

func (s *Server) EndAuction() {
	s.this.done = true
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

func max(this int32, that int32) int32 {
	if this < that {
		return that
	}
	return this
}

func formatAddress(port int) string {
	address := fmt.Sprintf("localhost:%v", port)
	return address
}

func (s *Server) CheckStatus(ctx context.Context, request *Replica.GetStatusRequest) (*Replica.StatusResponse, error) {
	var highestId int32
	for _, server := range s.allServers {
		highestId = max(int32(highestId), int32(server.id))
	}
	response := Replica.StatusResponse{ServerId: int32(highestId + 1)}
	return &response, nil
}

// func (s *Server) askNextReplica() {
// 	ctx := context.Background()
// 	address := fmt.Sprintf("localhost:%v", s.nextReplicaPort)
// 	conn, err := grpc.Dial(address, grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatalf("failed to connect to: %s", strconv.FormatInt(int64(s.port), 10))
// 	}

// 	nextRep := auction.NewAuctionServiceClient(conn)

// 	if _, err := nextRep.askNextReplica(); err != nil {
// 		log.Println(err)
// 	} else {
// 		log.Println("No errors")
// 	}
// }
