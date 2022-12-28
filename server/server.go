package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	bully "github.com/Mlth/Bully/proto"
	"google.golang.org/grpc"
)

type bullyServer struct {
	bully.BullyServer
}

var id int32
var isLeader bool
var otherServers []bully.BullyClient
var responseChan chan bool = make(chan bool)
var leaderId int32

func main() {
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000
	ownPortStr := strconv.Itoa(int(ownPort))
	arg2, _ := strconv.ParseInt(os.Args[2], 10, 32)
	numberOfReplicas := int32(arg2) + 9080

	list, err := net.Listen("tcp", ":"+ownPortStr)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", ownPortStr, err)
	}
	grpcServer := grpc.NewServer()
	bully.RegisterBullyServer(grpcServer, &bullyServer{})
	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("Failed to serve %v", err)
		}
	}()

	id = int32(arg1)
	if ownPort == 5000 {
		isLeader = true
	}

	otherServers = make([]bully.BullyClient, numberOfReplicas)
	for index := range otherServers {
		if index != int(id) {
			var conn *grpc.ClientConn
			var port int = index + 5000
			portStr := strconv.Itoa((port))

			conn, err := grpc.Dial(":"+portStr, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("Could not connect: %s", err)
			}
			defer conn.Close()

			c := bully.NewBullyClient(conn)
			otherServers[index] = c
		}
	}
	leaderId = 0

	if isLeader {
		doLeaderStuff()
	} else {
		doFollowerStuff()
	}
}

func doLeaderStuff() {
	log.Println("I am the leader")
	for {

	}
}

func doFollowerStuff() {
	for {
		go func() {
			log.Println("Checking on leader")
			response, _ := otherServers[leaderId].CheckLeaderConn(context.Background(), &bully.CheckMessage{})
			if response == nil {
				log.Println("Leader sent back nil value")
				localCheckHigherServers()
			} else {
				responseChan <- true
			}
		}()
		select {
		case <-time.After(5 * time.Second):
			log.Println("Leader did not respond in time")
			localCheckHigherServers()
		case <-responseChan:
			time.Sleep(10 * time.Second)
		}
	}
}

func localCheckHigherServers() {
	gotResponse := false
	for i := id; int(i) < len(otherServers); i++ {
		if otherServers[i] != nil {
			response, _ := otherServers[i].CheckForHigherServers(context.Background(), &bully.HigherServersMessage{})
			if response != nil {
				gotResponse = true
			}
		}
	}
	if !gotResponse {
		for _, server := range otherServers {
			server.NewCoordinator(context.Background(), &bully.CoordinaterMessage{Id: id})
		}
	}
}

func (s *bullyServer) NewCoordinator(ctx context.Context, in *bully.CoordinaterMessage) (*bully.CoordinaterAckMessage, error) {
	leaderId = in.Id
	return &bully.CoordinaterAckMessage{}, nil
}

func (s *bullyServer) CheckForHigherServers(ctx context.Context, in *bully.HigherServersMessage) (*bully.HigherServersReturnMessage, error) {
	return &bully.HigherServersReturnMessage{}, nil
}

func (s *bullyServer) checkLeaderConn(ctx context.Context, in *bully.CheckMessage) (*bully.CheckReturnMessage, error) {
	return &bully.CheckReturnMessage{}, nil
}
