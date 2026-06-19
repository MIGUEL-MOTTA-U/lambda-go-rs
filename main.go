package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"rs-lambda-go/internal/controller"
	"rs-lambda-go/internal/repository"
	"rs-lambda-go/internal/service"
)

const usersTableEnv = "USERS_TABLE"

func main() {
	tableName := strings.TrimSpace(os.Getenv(usersTableEnv))
	if tableName == "" {
		panic(fmt.Sprintf("missing required environment variable %s", usersTableEnv))
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("unable to load AWS config: %v", err))
	}

	db := dynamodb.NewFromConfig(cfg)
	userRepository := repository.NewDynamoUserRepository(db, tableName)
	userService := service.NewUserService(userRepository)
	userController := controller.NewUserController(userService)

	lambda.Start(userController.HandleRequest)
}
