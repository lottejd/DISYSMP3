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
	highestBid int32
}

func main() {

}

func (rep *Replica) updateBid(bid int32) {
	rep.highestBid = bid
	reportToPrimary()
}

func (rep *Replica) askNextReplica() {
	ctx := context.Background()
	address := fmt.Sprintf("localhost:%v", rep.nextReplicaPort)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to: %s", strconv.Itoa(node.port))
	}

	nextRep := new Replica(){}

	if _, err := nextRep.askNextReplica(); err != nil {
		log.Println(err)
	} else {
		log.Println("No errors")
	}
}

func pingPrimary() {
	for {

	}
}

func (rep *Replica) reportToPrimary() {
	
}