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

type ListingService interface {
	ListListings(ctx context.Context) ([]model.Listing, error)
	GetListing(ctx context.Context, id string) (model.Listing, error)
	CreateListing(ctx context.Context, listing model.Listing) (model.Listing, error)
	UpdateListing(ctx context.Context, id string, listing model.Listing) (model.Listing, error)
	DeleteListing(ctx context.Context, id string) error
}

type ListingController struct {
	service ListingService
}

func NewListingController(service ListingService) *ListingController {
	return &ListingController{service: service}
}

func (c ListingController) HandleRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	method := req.RequestContext.HTTP.Method
	path := normalizePath(req.RawPath)
	id := req.PathParameters["id"]

	switch {
	case method == http.MethodGet && path == "/listings":
		return c.listListings(ctx)
	case method == http.MethodGet && isListingIDPath(path):
		return c.getListing(ctx, idFromRequest(path, id))
	case method == http.MethodPost && path == "/listings":
		return c.createListing(ctx, req.Body)
	case method == http.MethodPut && isListingIDPath(path):
		return c.updateListing(ctx, idFromRequest(path, id), req.Body)
	case method == http.MethodDelete && isListingIDPath(path):
		return c.deleteListing(ctx, idFromRequest(path, id))
	case path == "/listings" || strings.HasPrefix(path, "/listings/"):
		return jsonResponse(http.StatusMethodNotAllowed, errorResponse{Message: "method not allowed"})
	default:
		return jsonResponse(http.StatusNotFound, errorResponse{Message: "route not found"})
	}
}

func (c ListingController) listListings(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	listings, err := c.service.ListListings(ctx)
	if err != nil {
		return serverError(err), nil
	}

	return jsonResponse(http.StatusOK, listings)
}

func (c ListingController) getListing(ctx context.Context, id string) (events.APIGatewayV2HTTPResponse, error) {
	listing, err := c.service.GetListing(ctx, id)
	if err != nil {
		return errorToListingResponse(err), nil
	}

	return jsonResponse(http.StatusOK, listing)
}

func (c ListingController) createListing(ctx context.Context, body string) (events.APIGatewayV2HTTPResponse, error) {
	listing, err := decodeListing(body)
	if err != nil {
		return jsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	createdListing, err := c.service.CreateListing(ctx, listing)
	if err != nil {
		return errorToListingResponse(err), nil
	}

	return jsonResponse(http.StatusCreated, createdListing)
}

func (c ListingController) updateListing(ctx context.Context, id string, body string) (events.APIGatewayV2HTTPResponse, error) {
	listing, err := decodeListing(body)
	if err != nil {
		return jsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	updatedListing, err := c.service.UpdateListing(ctx, id, listing)
	if err != nil {
		return errorToListingResponse(err), nil
	}

	return jsonResponse(http.StatusOK, updatedListing)
}

func (c ListingController) deleteListing(ctx context.Context, id string) (events.APIGatewayV2HTTPResponse, error) {
	if err := c.service.DeleteListing(ctx, id); err != nil {
		return errorToListingResponse(err), nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNoContent,
		Headers:    defaultHeaders(),
	}, nil
}

func decodeListing(body string) (model.Listing, error) {
	if strings.TrimSpace(body) == "" {
		return model.Listing{}, errors.New("request body is required")
	}

	var listing model.Listing
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&listing); err != nil {
		return model.Listing{}, fmt.Errorf("invalid json body: %w", err)
	}

	return listing, nil
}

func errorToListingResponse(err error) events.APIGatewayV2HTTPResponse {
	switch {
	case errors.Is(err, service.ErrInvalidListing):
		return responseFromError(http.StatusBadRequest, err)
	case errors.Is(err, repository.ErrListingNotFound):
		return responseFromError(http.StatusNotFound, err)
	case errors.Is(err, repository.ErrListingAlreadyExists):
		return responseFromError(http.StatusConflict, err)
	default:
		return serverError(err)
	}
}

func isListingIDPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 2 && parts[0] == "listings" && parts[1] != ""
}
