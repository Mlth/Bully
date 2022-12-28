package main

import (
	"log"
	"net"
	"os"
	"strconv"

	bully "github.com/Mlth/Bully/proto"
	"google.golang.org/grpc"
)

type bullyServer struct {
	bully.bullyServer
}

var id int32
var isLeader bool
var otherServers []bool
var leaderId int32

func main() {
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000
	ownPortStr := strconv.Itoa(int(ownPort))
	arg2, _ := strconv.ParseInt(os.Args[2], 10, 32)
	numberOfReplicas := int32(arg2) + 9080

	id = int32(arg1)
	if ownPort == 5000 {
		isLeader = true
	}
	otherServers = make([]bool, numberOfReplicas)
	for _, server := range otherServers {
		server = true
	}
	leaderId = 0

	list, err := net.Listen("tcp", ":"+ownPortStr)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", ownPortStr, err)
	}
	grpcServer := grpc.NewServer()
	bully.RegisterReplicationServer(grpcServer, &repServer{})
	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}

	if isLeader {
		doLeaderStuff()
	} else {
		doFollowerStuff()
	}
}

func doLeaderStuff() {
	log.Println("I am the leader")
}

func doFollowerStuff() {

}
