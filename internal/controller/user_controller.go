package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"rs-lambda-go/internal/model"
	"rs-lambda-go/internal/repository"
	"rs-lambda-go/internal/service"
)

type UserService interface {
	ListUsers(ctx context.Context) ([]model.User, error)
	GetUser(ctx context.Context, id string) (model.User, error)
	CreateUser(ctx context.Context, user model.User) (model.User, error)
	UpdateUser(ctx context.Context, id string, user model.User) (model.User, error)
	DeleteUser(ctx context.Context, id string) error
}

type UserController struct {
	service UserService
}

type errorResponse struct {
	Message string `json:"message"`
}

func NewUserController(service UserService) *UserController {
	return &UserController{service: service}
}

func (c UserController) HandleRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	method := req.RequestContext.HTTP.Method
	path := normalizePath(req.RawPath)
	id := req.PathParameters["id"]

	switch {
	case method == http.MethodGet && path == "/users":
		return c.listUsers(ctx)
	case method == http.MethodGet && isUserIDPath(path):
		return c.getUser(ctx, idFromRequest(path, id))
	case method == http.MethodPost && path == "/users":
		return c.createUser(ctx, req.Body)
	case method == http.MethodPut && isUserIDPath(path):
		return c.updateUser(ctx, idFromRequest(path, id), req.Body)
	case method == http.MethodDelete && isUserIDPath(path):
		return c.deleteUser(ctx, idFromRequest(path, id))
	case path == "/users" || strings.HasPrefix(path, "/users/"):
		return jsonResponse(http.StatusMethodNotAllowed, errorResponse{Message: "method not allowed"})
	default:
		return jsonResponse(http.StatusNotFound, errorResponse{Message: "route not found"})
	}
}

func (c UserController) listUsers(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	users, err := c.service.ListUsers(ctx)
	if err != nil {
		return serverError(err), nil
	}

	return jsonResponse(http.StatusOK, users)
}

func (c UserController) getUser(ctx context.Context, id string) (events.APIGatewayV2HTTPResponse, error) {
	user, err := c.service.GetUser(ctx, id)
	if err != nil {
		return errorToResponse(err), nil
	}

	return jsonResponse(http.StatusOK, user)
}

func (c UserController) createUser(ctx context.Context, body string) (events.APIGatewayV2HTTPResponse, error) {
	user, err := decodeUser(body)
	if err != nil {
		return jsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	createdUser, err := c.service.CreateUser(ctx, user)
	if err != nil {
		return errorToResponse(err), nil
	}

	return jsonResponse(http.StatusCreated, createdUser)
}

func (c UserController) updateUser(ctx context.Context, id string, body string) (events.APIGatewayV2HTTPResponse, error) {
	user, err := decodeUser(body)
	if err != nil {
		return jsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	updatedUser, err := c.service.UpdateUser(ctx, id, user)
	if err != nil {
		return errorToResponse(err), nil
	}

	return jsonResponse(http.StatusOK, updatedUser)
}

func (c UserController) deleteUser(ctx context.Context, id string) (events.APIGatewayV2HTTPResponse, error) {
	if err := c.service.DeleteUser(ctx, id); err != nil {
		return errorToResponse(err), nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNoContent,
		Headers:    defaultHeaders(),
	}, nil
}

func decodeUser(body string) (model.User, error) {
	if strings.TrimSpace(body) == "" {
		return model.User{}, errors.New("request body is required")
	}

	var user model.User
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&user); err != nil {
		return model.User{}, fmt.Errorf("invalid json body: %w", err)
	}

	return user, nil
}

func errorToResponse(err error) events.APIGatewayV2HTTPResponse {
	switch {
	case errors.Is(err, service.ErrInvalidUser):
		return responseFromError(http.StatusBadRequest, err)
	case errors.Is(err, repository.ErrUserNotFound):
		return responseFromError(http.StatusNotFound, err)
	case errors.Is(err, repository.ErrUserAlreadyExists):
		return responseFromError(http.StatusConflict, err)
	default:
		return serverError(err)
	}
}

func responseFromError(statusCode int, err error) events.APIGatewayV2HTTPResponse {
	resp, _ := jsonResponse(statusCode, errorResponse{Message: err.Error()})
	return resp
}

func jsonResponse(statusCode int, payload any) (events.APIGatewayV2HTTPResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return serverError(err), nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers:    defaultHeaders(),
		Body:       string(body),
	}, nil
}

func serverError(err error) events.APIGatewayV2HTTPResponse {
	body, _ := json.Marshal(errorResponse{Message: err.Error()})
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusInternalServerError,
		Headers:    defaultHeaders(),
		Body:       string(body),
	}
}

func defaultHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
		"Access-Control-Allow-Origin":  "*",
		"Content-Type":                 "application/json",
	}
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	if len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}
	return path
}

func isUserIDPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 2 && parts[0] == "users" && parts[1] != ""
}

func idFromRequest(path string, pathParameterID string) string {
	if strings.TrimSpace(pathParameterID) != "" {
		return pathParameterID
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 2 {
		return parts[1]
	}

	return ""
}
