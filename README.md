# rs-lambda-go

A Go-based AWS Lambda function that provides a serverless RESTful interface to manage user and listing data in Amazon DynamoDB tables. It handles request payload routing, validations, and database CRUD operations.

## Description

The `rs-lambda-go` application is an AWS Lambda function designed to handle HTTP request events forwarded by Amazon API Gateway V2. It implements two RESTful services:
1. **User Management**: Creation, retrieval, modification, and deletion operations on a DynamoDB database.
2. **Listing Management**: Creation, retrieval, modification, and deletion operations on real estate listing objects stored in a separate DynamoDB table.

The application automatically validates incoming HTTP request payloads, ensures required attributes are present, and returns standardized responses with CORS-compliant headers.

- For **Users**: When created without a pre-defined identifier, the system generates a cryptographically secure 16-byte random hexadecimal identifier.
- For **Listings**: When created without a pre-defined identifier, the system generates a cryptographically secure 16-byte random hexadecimal identifier. It also supports listing IDs as string or integer values, unmarshaling both automatically.

## Features

- **AWS Lambda Integration**: Formatted to handle `events.APIGatewayV2HTTPRequest` inputs and return `events.APIGatewayV2HTTPResponse` outputs using the AWS Lambda Go SDK.
- **DynamoDB CRUD Operations**: Communicates with DynamoDB to retrieve, insert, scan, update, and delete entries via the AWS SDK for Go V2.
- **Request Routing**: Houses an in-app router matching the HTTP method and path for `/users`, `/users/{id}`, `/listings`, and `/listings/{id}` resource endpoints.
- **CORS Support**: Serves default response headers including `Access-Control-Allow-Origin: *`, `Access-Control-Allow-Methods`, and `Access-Control-Allow-Headers`.
- **Validation**:
  - Enforces existence and formats for mandatory user attributes (`name`, `email`, `username`, `birthdate`).
  - Enforces existence and formats for listing attributes (`slug`, `url`, `language`, `title`, `property_type`, `operation_type`, `publication_status`).
- **Flexible ID Handling**: The listing service automatically accepts `listing_id` represented as a string or number, normalizes it, and maps it to a database-compliant format.

## Architecture

The project is structured according to a layered architecture pattern:

- **Entry Point (`main.go`)**: Initializes dependencies (AWS DynamoDB client, repositories, services, and controllers) and routes incoming HTTP requests to either the User controller or Listing controller based on the URL prefix.
- **Controller Layer (`internal/controller/`)**: Validates HTTP routes/methods, deserializes incoming JSON data, hands execution to the service layer, and maps service/repository errors to HTTP status codes.
- **Service Layer (`internal/service/`)**: Implements business and validation rules (e.g., verifying fields, appending timestamps, generating identifiers).
- **Repository Layer (`internal/repository/`)**: Defines database interactions. Maps model requests to DynamoDB API calls (`Scan`, `GetItem`, `PutItem`, `DeleteItem`).
- **Data Model (`internal/model/`)**: Declares the data models (`User` and `Listing`) mapping Go structures to both JSON payloads and DynamoDB attribute values.

## Technologies Used

| Category | Technology |
| :--- | :--- |
| Programming Language | Go (v1.26.4) |
| Core Framework | AWS Lambda Go (`github.com/aws/aws-lambda-go`) |
| SDK | AWS SDK for Go V2 (`github.com/aws/aws-sdk-go-v2`) |
| Database | Amazon DynamoDB |

## Requirements

- Go version `1.26.4` or newer.
- Two AWS DynamoDB tables:
  1. A users table configured with a string-type partition key named `id`.
  2. A listings table configured with a string-type partition key named `listing_id`.

## Installation

Download the required Go module dependencies:

```bash
go mod download
```

To compile the application as a binary named `bootstrap` (the standard entrypoint file name for custom runtime environments like `provided.al2023` in AWS Lambda) and package it into a zip archive:

### On Linux/macOS
```bash
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
zip lambda-handler.zip bootstrap
```

### On Windows (PowerShell)
```powershell
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o bootstrap main.go
Compress-Archive -Path bootstrap -DestinationPath lambda-handler.zip -Force
```

## Configuration

The application is configured using the following environment variables:

| Variable | Required | Description |
| :--- | :--- | :--- |
| `USERS_TABLE` | Yes | The name of the Amazon DynamoDB table where user records are stored. |
| `LISTINGS_TABLE` | Yes | The name of the Amazon DynamoDB table where listing records are stored. |

## Execution

The compiled binary runs inside the AWS Lambda environment upon receiving events from Amazon API Gateway V2.

To execute the Go process directly (e.g., for local testing tools that invoke binaries), ensure the environment variables are defined:

```bash
USERS_TABLE=users LISTINGS_TABLE=listings ./bootstrap
```

## Usage

### Endpoint Specifications

| Action | HTTP Method | Path | Request Body | Expected Response Status |
| :--- | :--- | :--- | :--- | :--- |
| List all users | `GET` | `/users` | None | `200 OK` |
| Get user by ID | `GET` | `/users/{id}` | None | `200 OK` |
| Create a user | `POST` | `/users` | User JSON | `201 Created` |
| Update a user | `PUT` | `/users/{id}` | User JSON | `200 OK` |
| Delete a user | `DELETE` | `/users/{id}` | None | `204 No Content` |
| List all listings | `GET` | `/listings` | None | `200 OK` |
| Get listing by ID | `GET` | `/listings/{id}` | None | `200 OK` |
| Create a listing | `POST` | `/listings` | Listing JSON | `201 Created` |
| Update a listing | `PUT` | `/listings/{id}` | Listing JSON | `200 OK` |
| Delete a listing | `DELETE` | `/listings/{id}` | None | `204 No Content` |

### JSON Payloads

#### Create Listing Request (`POST /listings`)
```json
{
  "slug": "c21-apartment-bogota",
  "url": "https://example.com/listings/c21-apartment-bogota",
  "language": "es",
  "title": "Apartamento en Bogotá",
  "property_type": "apartment",
  "subtype": "standard",
  "operation_type": "sale",
  "publication_status": "active",
  "location": {
    "country": "Colombia",
    "state": "Bogota",
    "city": "Bogota",
    "neighborhood": "Chicó",
    "address": "Calle 100",
    "stratum": 5,
    "coordinates": { "lat": 4.678, "lng": -74.048 }
  },
  "pricing": {
    "sale_price": 500000000,
    "rent_price": 0,
    "admin_fee": 300000,
    "taxes": 1500000,
    "currency": "COP",
    "display_price_text": "$500,000,000 COP"
  },
  "areas": {
    "land_area_m2": 0,
    "built_area_m2": 85,
    "private_area_m2": 80,
    "lot_area_m2": 0,
    "front_m": 0,
    "back_m": 0
  },
  "layout": {
    "bedrooms": 3,
    "bathrooms": 2,
    "half_bathrooms": 1,
    "parking_spaces": 2,
    "floors": 1,
    "unit_floor": 4
  },
  "structure": {
    "year_built": 2018,
    "age_years": 8,
    "construction_quality": "excellent",
    "conservation_status": "excellent",
    "terrain_type": "flat",
    "structure_type": "reinforced_concrete",
    "built_levels": 5
  },
  "features": {
    "indoor": ["elevator", "balcony"],
    "outdoor": ["gym", "security_24_7"],
    "commercial": [],
    "project": []
  },
  "media": {
    "photos": ["https://example.com/img1.jpg"],
    "photo_count": 1,
    "has_map": true,
    "has_video": false,
    "has_floorplans": true,
    "has_virtual_tour_360": false
  },
  "commercial": {
    "agent_name": "Juan Perez",
    "office_name": "C21 Colombia Centro",
    "phone": "+573000000000",
    "email": "juan.perez@c21colombia.com",
    "whatsapp_link": "https://wa.me/573000000000",
    "office_hours": "9:00 - 18:00"
  },
  "metadata": {
    "updated_at": "2026-06-19T20:00:00Z",
    "updated_age_text": "Updated recently",
    "breadcrumbs": ["Colombia", "Bogota", "Venta", "Apartamento"],
    "source_system": "century21colombia"
  }
}
```

## Project Structure

```text
rs-lambda-go/
├── internal/
│   ├── controller/
│   │   ├── listing_controller.go
│   │   └── user_controller.go
│   ├── model/
│   │   ├── listing.go
│   │   └── user.go
│   ├── repository/
│   │   ├── dynamo_listing_repository.go
│   │   ├── dynamo_user_repository.go
│   │   ├── listing_repository.go
│   │   └── user_repository.go
│   └── service/
│       ├── listing_service.go
│       └── user_service.go
├── .env
├── .gitignore
├── go.mod
├── go.sum
└── main.go
```

## Security

- **Payload Validation**: The service layer validates listing and user formats (required fields, value ranges, type validation).
- **Database Write Guard**: The database repositories use conditional expressions (`attribute_not_exists`) on `Create` and (`attribute_exists`) on `Delete` to prevent duplicate creations or deletion of non-existent keys.
- **Cross-Origin Resource Sharing (CORS)**: All controllers return HTTP headers allowing headers and methods standard to REST API specifications across any origin.
