package main

import (
	"fmt"
	"log"
	"os"
)

const (
	ServerPort    = 5000
	ServerLogFile = "serverLog"
	ConnectionNil = "TRANSIENT_FAILURE" // instead of nil when trying to connect to a port without a ReplicaService registered
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
	return fmt.Sprintf("Server id: %v, server port: %v, server is primary: %v,  server alive: %v,", s.id, s.port)
}
