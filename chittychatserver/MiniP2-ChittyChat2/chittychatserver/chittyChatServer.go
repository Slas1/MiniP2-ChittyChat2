package chittychatserver

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"
)

type messageUnit struct {
	ClientId    int32
	MessageBody string
	LampartTime int
}

type messageHandle struct {
	MQue []messageUnit
	mu   sync.Mutex
}

var messageHandleObject = messageHandle{}

var AllParticipents []int32
var LampartTime int32
var HighestId int32

type Server struct {
}

func (is *Server) JoinServer(ctx context.Context, req *FromClient) (*FromServer, error) {
	var clientNumber = HighestId + 1
	HighestId++
	log.Printf("Participant %d  joined Chitty-Chat at Lamport time %d", clientNumber, LampartTime)
	LampartTime++
	AllParticipents = append(AllParticipents, clientNumber)
	return &FromServer{ID: clientNumber, Body: "Participant " + strconv.Itoa(int(clientNumber)) + " joined Chitty-Chat at Lamport time " + strconv.Itoa(int(LampartTime-1)), Time: LampartTime - 1}, nil
}

func (is *Server) LeaveServer(ctx context.Context, req *FromClient) (*FromServer, error) {
	log.Printf("Participant %d left Chitty-Chat at Lamport time %d", req.ID, LampartTime)
	LampartTime++
	var stadin []int32
	for i := 0; i < len(AllParticipents); i++ {
		if AllParticipents[i] != req.ID {
			stadin = append(stadin, AllParticipents[i])
		}
	}
	AllParticipents = stadin
	return &FromServer{ID: req.ID, Body: "Participant " + strconv.Itoa(int(req.ID)) + " left Chitty-Chat at Lamport time " + strconv.Itoa(int(LampartTime-1)), Time: LampartTime - 1}, nil
}

func (is *Server) ChatService(csi Services_ChatServiceServer) error {

	//clientUniqueCode := rand.Intn(1e6)
	errch := make(chan error)

	// receive messages - init a go routine
	go receiveFromStream(csi, errch)

	// send messages - init a go routine
	go sendToStream(csi, errch)

	return <-errch

}

//Receive messages
func receiveFromStream(csi_ Services_ChatServiceServer, errch_ chan error) {
	//implement a loop
	for {
		mssg, err := csi_.Recv()
		if err != nil {
			log.Printf("Error in receiving message from client :: %v", err)
			errch_ <- err
		} else {

			messageHandleObject.mu.Lock()

			messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
				ClientId:    mssg.ID,
				MessageBody: mssg.Body,
				LampartTime: int(mssg.Time),
			})

			if LampartTime < int32(mssg.Time) {
				LampartTime = mssg.Time
			}

			log.Printf("%v", messageHandleObject.MQue[len(messageHandleObject.MQue)-1])

			messageHandleObject.mu.Unlock()
		}
	}
}

//Send Messages
func sendToStream(csi_ Services_ChatServiceServer, errch_ chan error) {
	//implement a loop
	for {

		//loop through messages in MQue
		for {

			time.Sleep(500 * time.Millisecond)

			messageHandleObject.mu.Lock()

			if len(messageHandleObject.MQue) == 0 {
				messageHandleObject.mu.Unlock()
				break
			}

			senderID := messageHandleObject.MQue[0].ClientId
			messageFromClient := messageHandleObject.MQue[0].MessageBody
			timeOfMessage := messageHandleObject.MQue[0].LampartTime

			messageHandleObject.mu.Unlock()
			//send message to designated client (do not send to the same client)

			for i := 0; i < len(AllParticipents); i++ {
				err := csi_.Send(&FromServer{ID: senderID, Body: messageFromClient, Time: int32(timeOfMessage)})

				if err != nil {
					errch_ <- err
				}

				messageHandleObject.mu.Lock()

				if len(messageHandleObject.MQue) > 1 {
					messageHandleObject.MQue = messageHandleObject.MQue[1:] // delete the message at index 0 after sending to receiver
				} else {
					messageHandleObject.MQue = []messageUnit{}
				}

				messageHandleObject.mu.Unlock()
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
