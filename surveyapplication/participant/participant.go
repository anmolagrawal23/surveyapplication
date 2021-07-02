package main

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	surveyservice "github.com/anmolagrawal23/surveyapplication/proto"
	"google.golang.org/grpc"
)

var surveyID uint32

var wg = sync.WaitGroup{}

func main() {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "localhost:7777", grpc.WithInsecure())
	if err != nil {
		log.Println("Error connecting server:", err)
	}
	defer conn.Close()

	client := surveyservice.NewSurveyServiceClient(conn)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go SendResponse(ctx, client)
	}
	wg.Wait()
}

func SendResponse(ctx context.Context, client surveyservice.SurveyServiceClient) {

	log.Println("Starting to send the responses")
	stream, err := client.SendResponse(ctx)
	if err != nil {
		log.Fatal("Error creating the stream to send questions.", err)
	}

	for _, ques := range getResponseList() {
		stream.Send(&ques)
		time.Sleep(1 * time.Second)
	}
	_, errClose := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("Error closing the stream to send questions.", errClose)
	}

	wg.Done()
}

func getResponseList() []surveyservice.Response {

	var ResponseList = []surveyservice.Response{
		surveyservice.Response{
			SurveyID:    1,
			QuestionNum: 1,
			Response:    strconv.Itoa(rand.Intn(4) + 1),
		},
		surveyservice.Response{
			SurveyID:    1,
			QuestionNum: 2,
			Response:    strconv.Itoa(rand.Intn(4) + 1),
		},
		surveyservice.Response{
			SurveyID:    1,
			QuestionNum: 3,
			Response:    strconv.Itoa(rand.Intn(4) + 1),
		},
	}
	return ResponseList
}
