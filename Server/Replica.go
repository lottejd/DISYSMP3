package main

import (
	"context"
	"fmt"
	"github.com/lottejd/DISYSMP3/Auction"
	"google.golang.org/grpc"
	"log"
	"strconv"
)

type Replica struct {
	id              int32
	primary         bool
	port            int32
	nextReplicaPort int32
}

func main() {

}

func askNextReplica(rep *Replica) {
	ctx := context.Background()
	address := fmt.Sprintf("localhost:%v", rep.nextReplicaPort)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to: %s", strconv.Itoa(node.port))
	}
	nextNode := mutex.NewMutexServiceClient(conn)

	if _, err := nextNode.Token(ctx, &mutex.EmptyRequest{}); err != nil {
		log.Println(err)
	} else {
		log.Println("No errors")
	}
}

func pingPrimary() {
	for {

	}
}
