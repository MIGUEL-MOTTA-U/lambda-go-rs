package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"rs-lambda-go/internal/controller"
	"rs-lambda-go/internal/repository"
	"rs-lambda-go/internal/service"
)

const (
	usersTableEnv    = "USERS_TABLE"
	listingsTableEnv = "LISTINGS_TABLE"
)

type Router struct {
	userController    *controller.UserController
	listingController *controller.ListingController
}

func (r Router) Route(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	path := strings.TrimRight(req.RawPath, "/")
	if path == "/users" || strings.HasPrefix(path, "/users/") {
		return r.userController.HandleRequest(ctx, req)
	}
	if path == "/listings" || strings.HasPrefix(path, "/listings/") {
		return r.listingController.HandleRequest(ctx, req)
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 404,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
			"Access-Control-Allow-Origin":  "*",
			"Content-Type":                 "application/json",
		},
		Body: `{"message":"route not found"}`,
	}, nil
}

func main() {
	usersTable := strings.TrimSpace(os.Getenv(usersTableEnv))
	if usersTable == "" {
		panic(fmt.Sprintf("missing required environment variable %s", usersTableEnv))
	}

	listingsTable := strings.TrimSpace(os.Getenv(listingsTableEnv))
	if listingsTable == "" {
		panic(fmt.Sprintf("missing required environment variable %s", listingsTableEnv))
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("unable to load AWS config: %v", err))
	}

	db := dynamodb.NewFromConfig(cfg)

	userRepo := repository.NewDynamoUserRepository(db, usersTable)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	listingRepo := repository.NewDynamoListingRepository(db, listingsTable)
	listingService := service.NewListingService(listingRepo)
	listingController := controller.NewListingController(listingService)

	router := Router{
		userController:    userController,
		listingController: listingController,
	}

	lambda.Start(router.Route)
}
