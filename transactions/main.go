package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// TransactionType is a custom type for representing the transaction type.
type TransactionType string

const (
	CREDIT TransactionType = "CREDIT"
	DEBIT  TransactionType = "DEBIT"
)

var dynamoDBClient *dynamodb.Client
var tableName string

type Transaction struct {
	ID        string          `json:"id"               dynamodbav:"id"`
	UserID    string          `json:"userId"           dynamodbav:"userId"`
	Origin    string          `json:"origin"           dynamodbav:"origin"`
	Timestamp int64           `json:"timeStamp"        dynamodbav:"timeStamp"`
	Amount    float64         `json:"amount"           dynamodbav:"amount"`
	Type      TransactionType `json:"transactionType"  dynamodbav:"transactionType"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Prepare data using the helper library
	item := Transaction{
		ID:        uuid.NewString(),
		UserID:    "1",
		Timestamp: time.Now().Unix(),
		Type:      CREDIT,
		Amount:    500,
	}

	// Marshal the item into a map of AttributeValues
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName), // replace with your table name
	}
	// Put a transaction item into DynamoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // The cancel should be deferred so resources are freed up

	_, err = dynamoDBClient.PutItem(ctx, input)
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

func createDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %w", err)
	}

	return dynamodb.NewFromConfig(cfg), nil
}

func init() {
	// Fetch DynamoDB table name from environment variable
	tableName = os.Getenv("TABLE_NAME")
	if tableName == "" {
		panic("DYNAMO_TABLE_NAME environment variable not set")
	}

	ctx := context.Background()
	var err error
	dynamoDBClient, err = createDynamoDBClient(ctx)

	if err != nil {
		panic(fmt.Errorf("error creating DynamoDB client: %v", err))
	}
}

func main() {
	lambda.Start(handler)
}
