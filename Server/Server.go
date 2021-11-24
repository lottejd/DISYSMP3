package main

import (
	"context"

	"github.com/lottejd/DISYSMP3/Replica"
)

type Server struct {
	Replica.UnimplementedReplicaServiceServer
	id   int32
	port int32
}

func main() {

	//init
	_, conn := Connect(ServerPort) //implement port increment
	server := CreateReplica(conn)

	//setup listen on port
	StartReplicaService(server.port, &server)
}

// gRPC service sends replicas own status
func (s *Server) CheckStatus(ctx context.Context, empty *Replica.EmptyRequest) (*Replica.StatusResponse, error) {
	response := Replica.StatusResponse{ServerId: s.id, Msg: "Alive and well"}
	return &response, nil
}
