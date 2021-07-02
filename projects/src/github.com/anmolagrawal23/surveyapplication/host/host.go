package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/go-echarts/go-echarts/v2/charts"

	surveyservice "github.com/anmolagrawal23/surveyapplication/proto"
	"google.golang.org/grpc"
)

var surveyID uint32

var surveyDesc = "Survey to gather project reviews."

type Response struct {
	queNum   uint32
	Response int
}

var response = make(chan Response, 100)

var queCount [5]int

func main() {

	go func() {
		log.Println("Starting dashboard url on localhost:8081")
		http.HandleFunc("/", Httpserver)
		http.ListenAndServe(":8081", nil)
	}()

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "localhost:7777", grpc.WithInsecure())
	if err != nil {
		log.Println("Error connecting server:", err)
	}
	defer conn.Close()

	client := surveyservice.NewSurveyServiceClient(conn)
	CreateSurvey(ctx, client)
	SendQuestions(ctx, client)
	time.Sleep(10 * time.Second)
	ReceiveResponses(ctx, client, surveyID)
	//close(response)
	//log.Println("Local channel data:")
	//for res := range response {
	//	log.Println(res)
	//}
}

func CreateSurvey(ctx context.Context, client surveyservice.SurveyServiceClient) {
	log.Println("Creating a new survey...")
	survey := surveyservice.SurveyDesc{
		Description: surveyDesc,
	}
	res, err := client.CreateSurvey(ctx, &survey)
	if err != nil {
		log.Fatal("Error creating the survey.", err)
	}
	surveyID = res.SurveyID
	log.Println("Survey created with ID: ", surveyID)
}

func SendQuestions(ctx context.Context, client surveyservice.SurveyServiceClient) {

	log.Println("Starting to send the questions")
	stream, err := client.SendQuestions(ctx)
	if err != nil {
		log.Fatal("Error creating the stream to send questions.", err)
	}
	questionList := getQuestionList(surveyID)
	for _, ques := range questionList {
		stream.Send(&ques)
	}
	_, errClose := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("Error closing the stream to send questions.", errClose)
	}
}

func ReceiveResponses(ctx context.Context, client surveyservice.SurveyServiceClient, surveyID uint32) {

	log.Println("Starting a stream to receive responses...")
	stream, err := client.ReceiveResponse(ctx, &surveyservice.SurveyID{SurveyID: surveyID})
	if err != nil {
		log.Fatal("Error creating receiving stream.")
	}

	done := time.NewTicker(120 * time.Second)
	timer := time.NewTicker(2 * time.Second)

	for {
		select {

		case <-done.C:
			err := stream.CloseSend()
			if err != nil {
				log.Fatal("Error closing the receiving stream:", err)
			}
			log.Println("Closing receiving stream.")
			return
		case <-timer.C:

			res, err := stream.Recv()
			if err != nil {
				log.Fatal("Error receiving the response from stream.")
			}
			log.Println("Received response", res)

			resList := strings.Split(res.Response, ",")
			var resIntList []int

			for _, v := range resList {
				intValue, _ := strconv.Atoi(v)
				resIntList = append(resIntList, intValue)
			}
			for _, resInt := range resIntList {
				log.Printf("Pushing response {%v,%v} to local channel..", res.QuestionNum, resInt)
				response <- Response{
					queNum:   res.QuestionNum,
					Response: resInt,
				}
			}
		}

	}
}

func generateBarItems() []opts.BarData {

	log.Println("Generating bar items...")
	items := make([]opts.BarData, 0)
	//done := time.NewTicker(2 * time.Second)
	for {
		select {
		case res := <-response:
			if res.queNum == 1 {
				queCount[res.Response]++
			}
		default:
			log.Println(queCount)
			for i := 1; i < 5; i++ {
				items = append(items, opts.BarData{Value: queCount[i]})
			}
			log.Println("Returning bar items", items)
			return items
		}
	}
	return items
}

func Httpserver(w http.ResponseWriter, r *http.Request) {

	log.Println("Request to show the graph...")
	// create a new line instance
	fmt.Fprintf(w, "<head><meta http-equiv=\"refresh\" content=\"2\"></head>")
	fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", "Golang Chat Example", "Line Chart")
	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    surveyDesc,
		Subtitle: "How was the presentation?",
	}))

	bar.SetXAxis([]string{"Poor", "Average", "Good", "Excellent"}).
		AddSeries("Count", generateBarItems())

	bar.Render(w)
}

func getQuestionList(surveyID uint32) []surveyservice.Question {

	var QuestionsList = []surveyservice.Question{
		surveyservice.Question{
			SurveyID:    surveyID,
			QuestionNum: 1,
			Description: "How was the presentation?",
			Type:        surveyservice.Question_MCQ,
			Graph:       surveyservice.Question_BAR,
			Options: []*surveyservice.Option{
				&surveyservice.Option{
					OptionNum: 1,
					Option:    "Poor",
				},
				&surveyservice.Option{
					OptionNum: 2,
					Option:    "Average",
				},
				&surveyservice.Option{
					OptionNum: 3,
					Option:    "Good",
				},
				&surveyservice.Option{
					OptionNum: 4,
					Option:    "Excellent",
				},
			},
		},

		surveyservice.Question{
			SurveyID:    surveyID,
			QuestionNum: 2,
			Description: "How was the demo?",
			Type:        surveyservice.Question_MCQ,
			Graph:       surveyservice.Question_BAR,
			Options: []*surveyservice.Option{
				&surveyservice.Option{
					OptionNum: 1,
					Option:    "Poor",
				},
				&surveyservice.Option{
					OptionNum: 2,
					Option:    "Average",
				},
				&surveyservice.Option{
					OptionNum: 3,
					Option:    "Good",
				},
				&surveyservice.Option{
					OptionNum: 4,
					Option:    "Excellent",
				},
			},
		},
		surveyservice.Question{
			SurveyID:    surveyID,
			QuestionNum: 3,
			Description: "How was the session?",
			Type:        surveyservice.Question_MCQ,
			Graph:       surveyservice.Question_BAR,
			Options: []*surveyservice.Option{
				&surveyservice.Option{
					OptionNum: 1,
					Option:    "Poor",
				},
				&surveyservice.Option{
					OptionNum: 2,
					Option:    "Average",
				},
				&surveyservice.Option{
					OptionNum: 3,
					Option:    "Good",
				},
				&surveyservice.Option{
					OptionNum: 4,
					Option:    "Excellent",
				},
			},
		},
	}
	return QuestionsList
}
