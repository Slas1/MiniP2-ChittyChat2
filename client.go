package main

import (
	"bufio"
	"context"
	"fmt"
	"grpcChittyChatServer/chittychatserver/MiniP2-ChittyChat2/chittychatserver"
	"log"
	"os"
	"strings"

	"github.com/thecodeteam/goodbye"
	"google.golang.org/grpc"
)

var ID int32
var LampartTime int32

func main() {
	ctx := context.Background()

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not conect: %s", err)
	}

	defer conn.Close()
	client := chittychatserver.NewServicesClient(conn)

	//Skulle være en exithandler
	defer goodbye.Exit(ctx, -1)
	goodbye.Notify(ctx)
	goodbye.Register(func(ctx context.Context, sig os.Signal) {
		fmt.Println("Dette driller")
		LeaveServer(client, chittychatserver.FromClient{ID: ID, Body: "Im leaving", Time: LampartTime + 1})
	})


	//Joining server
	JoinServer(client, chittychatserver.FromClient{ID: -1, Body: "Im joining the server", Time: 0})

	//call ChatService to create a stream
	stream, err := client.ChatService(context.Background())
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)
	}

	// implement communication with gRPC server
	ch := clienthandle{stream: stream}
	//ch.clientConfig()
	go ch.sendMessage()
	go ch.receiveMessage()

	//blocker
	bl := make(chan bool)
	<-bl

}

type clienthandle struct {
	stream   chittychatserver.Services_ChatServiceClient
	clientID string
}

/*
func (ch *clienthandle) clientConfig() {

	reader := bufio.NewReader(os.Stdin)
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf(" Failed to read from console :: %v", err)
	}
	ch.clientName = strings.Trim(name, "\r\n")
}
*/

func JoinServer(c chittychatserver.ServicesClient, req chittychatserver.FromClient) {
	response, err := c.JoinServer(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error when calling JoinServer: %s", err)
	}

	ID = response.ID
	LampartTime = response.Time + 1

	log.Printf(response.Body)
}

func LeaveServer(c chittychatserver.ServicesClient, req chittychatserver.FromClient) {
	_, err := c.LeaveServer(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error when calling LeaveServer: %s", err)
	}
}

func (ch *clienthandle) sendMessage() {
	// create a loop
	for {
		reader := bufio.NewReader(os.Stdin)
		clientMessage, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf(" Failed to read from console :: %v", err)
		}
		clientMessage = strings.Trim(clientMessage, "\r\n")

		clientMessageBox := &chittychatserver.FromClient{
			ID:   ID,
			Body: clientMessage,
		}

		err = ch.stream.Send(clientMessageBox)

		if err != nil {
			log.Printf("Error while sending message to server :: %v", err)
		}

	}
}

//receive message
func (ch *clienthandle) receiveMessage() {
	for {
		mssg, err := ch.stream.Recv()
		if err != nil {
			log.Printf("Error in receiving message from server :: %v", err)
		}

		//print message to console
		fmt.Printf("%d : %s \n", mssg.ID, mssg.Body)

	}
}
