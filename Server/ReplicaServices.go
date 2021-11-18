package main

import (
	"fmt"

	"github.com/lottejd/DISYSMP3/Replica"
	"google.golang.org/grpc"
)

func EvalPrimary(conn *grpc.ClientConn) bool {
	return !IsConnectable(conn)
}



func IsConnectable(conn *grpc.ClientConn) bool {
	return conn.GetState().String() != ConnectionNil
}

func EvalServers(conn *grpc.ClientConn, replicaInfo *Replica.ReplicaInfo) map[int32]Server {
	serverMap := make(map[int32]Server)
	if IsConnectable(conn) {
		for _, replica := range replicaInfo.Replicas {
			serverMap[replica.GetServerId()] = *CreateProxyReplica(replica.GetServerId(), replica.GetPort())
		}
	}
	return serverMap
}

func (s *Server) ServerMapToReplicaInfoArray() []*Replica.ReplicaInfo {
	var servers []*Replica.ReplicaInfo
	for _, Server := range s.allServers {
		temp := Replica.ReplicaInfo{ServerId: Server.id, Port: Server.port}
		servers = append(servers, &temp)
	}
	return servers
}

func (s *Server) DisplayAllReplicas() {
	for _, server := range s.allServers {
		if server.alive {
			fmt.Println(server.ToString())
		}
	}
}


func (s *Server) FindServers() {
	for i := 0; i < 10; i++ {
		if int(s.id) == i {
			continue
		}
		serverId, conn := Connect(int32(ServerPort + i))
		if IsConnectable(conn) && s.allServers[int32(serverId)].port == 0 {
			replica := CreateProxyReplica(serverId, int32(ServerPort+i))
			fmt.Println(s.port)
			s.AddReplica(replica)
		}
	}
}