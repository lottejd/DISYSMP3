package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

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
func (s *Server) FindServersAndAddToMap(c *Counter) {
	// add counter *int and bool primarySeen if for 3 rounds primarySeen false start Election
	var primarySeen bool
	ctx := context.Background()
	if c.counter >= 3 {
		fmt.Println(c.counter)
		c.counter = 0
		s.isPrimaryDead <- true
	}
	for i := 0; i < 10; i++ {
		if int(s.id) == i {
			s.allServers[s.id] = *s
		} else {
			serverId, conn := Connect(int32(ServerPort + i))
			if IsConnectable(conn) {
				replica := CreateProxyReplica(serverId, int32(ServerPort+i))
				replicaClient, _ := ConnectToReplicaClient(replica.port)

				request := Replica.GetStatusRequest{ServerId: replica.id}
				response, _ := replicaClient.CheckStatus(ctx, &request)
				if response.Primary {
					primarySeen = true
					replica.SetPrimary()
				}
				s.AddReplicaToMap(replica)
				conn.Close()
			} else {
				// fmt.Printf("removing: %v from the map\n", i)
				s.RemoveReplicaFromMap(int32(i))
			}
		}
	}
	if primarySeen || s.primary {
		c.counter = 0
	} else {
		c.counter = c.counter + 1
	}
	ctx.Done()
}

type Counter struct {
	counter int
}

func (s *Server) RemoveReplicaFromMap(serverId int32) {
	delete(s.allServers, serverId)
}

func Listen(port int32, s *Server) {
	// start peer to peer service
	go StartReplicaService(port, s)

	// start auction service
	if s.primary {
		StartAuctionService(s)

	}
}

type Service struct {
	stopService chan bool
	lis         net.Listener
}

func (service *Service) Serve() {
	defer service.lis.Close()
	for {
		select {
		case <-service.stopService:
			fmt.Printf("Stop listing on port %v", service.lis.Addr())
			service.lis.Close()
		default:
			time.Sleep(time.Second * 1)
		}

	}
}

func StartReplicaService(port int32, s *Server) {
	// stopServiceChan := make(chan bool)
	lis, _ := net.Listen("tcp", FormatAddress(s.port))
	defer lis.Close()

	// service := Service{stopService: stopServiceChan, lis: lis}
	// s.replicaService = service
	// go s.replicaService.Serve()

	grpcServer := grpc.NewServer()
	Replica.RegisterReplicaServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		fmt.Printf("Failed to serve %v on port %v\n", s.id, s.port)
	}
	fmt.Println("reachable in Startreplicalisten")
}

func StartAuctionService(s *Server) {
	// stopServiceChan := make(chan bool)
	lis, _ := net.Listen("tcp", FormatAddress(ClientPort))
	defer lis.Close()

	// service := Service{stopService: stopServiceChan, lis: lis}
	// s.auctionService = service
	// go s.auctionService.Serve()

	grpcServer := grpc.NewServer()
	Auction.RegisterAuctionServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		fmt.Printf("Failed to serve %v on port %v\n", s.id, ClientPort)
	}
}
