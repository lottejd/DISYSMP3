package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

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
	return fmt.Sprintf("Server id: %v, server port: %v", s.id, s.port)
}

// connect to a replica with at max 3 retries
func ConnectToReplicaClient(conn *grpc.ClientConn) (Replica.ReplicaServiceClient, string) {
	ctx := context.Background()
	
	client := Replica.NewReplicaServiceClient(conn)
	ack, err := client.CheckStatus(ctx, &Replica.EmptyRequest{})
	if err != nil || ack.GetServerId() < 0 {
		return client, "failed"
	}

	ctx.Done()
	return client, "succes"
}

// connect to address
func Connect(port int) (*grpc.ClientConn, string) {
	address := fmt.Sprintf("localhost:%v", port)
	var status string
	var conn *grpc.ClientConn
	for i := 0; i < 3; i++ {
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			status = "failed"
			time.Sleep(time.Millisecond * 250)
			continue
		} else {
			return conn, "succes"
		}
	}
	return conn, status
}
