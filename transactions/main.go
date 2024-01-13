package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

var dynamoDBClient *dynamodb.DynamoDB
var tableName string

type Transaction struct {
	ID        string `json:"ID"`
	Timestamp int64  `json:"Timestamp"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Prepare data using the helper library
	item := Transaction{
		ID:        uuid.NewString(),
		Timestamp: time.Now().Unix(),
	}
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	// Put a transaction item into DynamoDB
	_, err = dynamoDBClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	// Return JSON encoding of successfully put item
	transaction, err := json.Marshal(item)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: http.StatusInternalServerError,
		}, nil
	}
	return events.APIGatewayProxyResponse{
		Body:       string(transaction),
		StatusCode: http.StatusOK,
	}, nil
}

func init() {
	// Fetch DynamoDB table name from environment variable
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		panic("DYNAMO_TABLE_NAME environment variable not set")
	}
	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		// Add your AWS credentials here if not using IAM roles in Lambda
		// Credentials: credentials.NewStaticCredentials("your-access-key-id", "your-secret-access-key", ""),
	})
	if err != nil {
		panic(fmt.Errorf("error creating AWS session: %v", err))
	}

	// Create a new DynamoDB client
	dynamoDBClient = dynamodb.New(sess)
}

func main() {
	lambda.Start(handler)
}
