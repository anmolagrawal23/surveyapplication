package server

import (
	"context"
	"io"
	"log"
	"sync/atomic"

	surveyservice "github.com/anmolagrawal23/surveyapplication/proto"
)

var SurveyID uint32

type Question struct {
	questionNum uint32
	description string
	quesType    surveyservice.Question_Type
	graphType   surveyservice.Question_Graph
	options     []*surveyservice.Option
}

type Response struct {
	queNumber uint32
	response  string
}

type Survey struct {
	surveyID    uint32
	description string
	questions   []Question
	responses   chan Response
}

var SurveyMap map[uint32]*Survey

type Server struct {
	surveyservice.UnimplementedSurveyServiceServer
}

func (s *Server) CreateSurvey(_ context.Context, surveyDesc *surveyservice.SurveyDesc) (*surveyservice.SurveyID, error) {

	log.Println("Received request to create a new survey...")
	atomic.AddUint32(&SurveyID, 1)
	sID := SurveyID
	log.Println("Generated Survey ID: ", sID)

	if SurveyMap == nil {
		SurveyMap = make(map[uint32]*Survey)
	}

	SurveyMap[sID] = &Survey{
		surveyID:    sID,
		description: surveyDesc.Description,
	}

	log.Println("Survey created: ", SurveyMap[sID])

	res := &surveyservice.SurveyID{
		SurveyID: sID,
	}
	return res, nil
}

func (s *Server) SendQuestions(stream surveyservice.SurveyService_SendQuestionsServer) error {

	log.Println("Starting stream to receive questions...")
	for {
		que, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				err := stream.SendAndClose(&surveyservice.EmptyMessage{})
				if err != nil {
					log.Fatal("Error closing the stream.", err)
				}
				return nil
			} else {
				log.Fatal("Error receiving on question stream: ", err)
				return err
			}
		}

		sID := que.SurveyID
		if _, ok := SurveyMap[sID]; !ok {
			log.Fatal("Survey not present for the ID: ", sID, err)
		}

		newQue := Question{
			questionNum: que.QuestionNum,
			description: que.Description,
			quesType:    que.Type,
			graphType:   que.Graph,
			options:     que.Options,
		}

		SurveyMap[sID].questions = append(SurveyMap[sID].questions, newQue)
		log.Printf("Added question %v for SurveyID %v", newQue, sID)

	}
	return nil
}

func (s *Server) SendResponse(stream surveyservice.SurveyService_SendResponseServer) error {

	log.Println("Starting stream to receive response...")
	for {
		res, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				err := stream.SendAndClose(&surveyservice.EmptyMessage{})
				if err != nil {
					log.Fatal("Error closing the stream.", err)
				}
				return nil
			} else {
				log.Fatal("Error receiving on question stream: ", err)
				return err
			}
		}

		sID := res.SurveyID
		if _, ok := SurveyMap[sID]; !ok {
			log.Fatal("Survey not present for the ID: ", sID, "Error:", err)
		} else {
			log.Println("Survey present for survey ID:", sID)
		}

		if SurveyMap[sID].responses == nil {
			SurveyMap[sID].responses = make(chan Response, 1000)
			log.Println("Assigning memory to the response channel.")
		}

		SurveyMap[sID].responses <- Response{
			queNumber: res.QuestionNum,
			response:  res.Response,
		}

		log.Printf("Response {%v,%v} received successfully.", res.QuestionNum, res.Response)
	}
	return nil
}

func (s *Server) ReceiveResponse(sID *surveyservice.SurveyID, stream surveyservice.SurveyService_ReceiveResponseServer) error {

	log.Println("Received request to send responses for survey ID:", sID.SurveyID)

	if _, ok := SurveyMap[sID.SurveyID]; !ok {
		log.Fatal("Survey not found for ID:", sID.SurveyID)
		// return errors.New("Survey not found for ID")
	}

	log.Println("Survey found for ID:", sID.SurveyID)

	for {
		select {
		case resVal := <-SurveyMap[sID.SurveyID].responses:
			res := surveyservice.Response{
				SurveyID:    sID.SurveyID,
				QuestionNum: resVal.queNumber,
				Response:    resVal.response,
			}
			err := stream.Send(&res)
			if err != nil {
				log.Fatal("Error sending response stream to host.")
				return err
			}
			log.Printf("Sent response to host: {%v,%v,%v}", sID.SurveyID, resVal.queNumber, resVal.response)

			//		case <-stream.Context().Done():
			//			return nil
		}
	}
	return nil
}
