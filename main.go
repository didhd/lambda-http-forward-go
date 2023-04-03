package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(ctx context.Context, request events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	return handleAPIRequest(request)
}

func handleAPIRequest(request events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	// Replace the URL with your target API URL
	apiURL := os.Getenv("APIURL")

	// Combine the base URL with the path and query string
	url := fmt.Sprintf("%s%s?%s", apiURL, request.Path, request.QueryStringParameters)

	// Prepare the request
	apiRequest, err := http.NewRequest(request.HTTPMethod, url, bytes.NewBuffer([]byte(request.Body)))
	if err != nil {
		return events.ALBTargetGroupResponse{StatusCode: http.StatusInternalServerError}, err
	}

	// Copy headers from the original request to the proxy request
	for key, value := range request.Headers {
		apiRequest.Header.Set(key, value)
	}

	// Send the request
	client := &http.Client{}
	response, err := client.Do(apiRequest)
	if err != nil {
		return events.ALBTargetGroupResponse{StatusCode: http.StatusInternalServerError}, err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return events.ALBTargetGroupResponse{StatusCode: http.StatusInternalServerError}, err
	}

	// Copy headers from the API response to the proxy response
	headers := make(map[string]string)
	for key, values := range response.Header {
		headers[key] = values[0]
	}

	// Return the API response as the Lambda proxy response
	return events.ALBTargetGroupResponse{
		StatusCode: response.StatusCode,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
