package KLaaS

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

var client *openai.Client

const prompt = "Write a captains log style entry as a real starship captain would based on the following diary entry"

func init() {
	openaiToken, ok := os.LookupEnv("OPENAI_TOKEN")
	if !ok {
		log.Panicf("ERROR - Could not retrieve OPENAI_TOKEN")
	}

	client = openai.NewClient(openaiToken)

	functions.HTTP("MainHandler", mainHandler)
}

type request struct {
	OriginalText string    `json:"original_text"`
	Date         time.Time `json:"date"`
}

type response struct {
	LogText string `json:"log_text"`
}

// mainHandler is an HTTP Cloud Function with a request parameter.
func mainHandler(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Could not read request")
		return
	}

	gptResp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("%s:\n\n%s", prompt, req.OriginalText),
				},
			},
		},
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Could not generate log entry: %v", err)
	}

	resp := response{LogText: gptResp.Choices[0].Message.Content}

	json.NewEncoder(w).Encode(resp)
}
