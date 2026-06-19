# rs-lambda-go

A Go-based AWS Lambda function that provides a serverless RESTful interface to manage user data in an Amazon DynamoDB table. It handles request payload routing, validations, and database CRUD operations.

## Description

The `rs-lambda-go` application is an AWS Lambda function designed to handle HTTP request events forwarded by Amazon API Gateway V2. It implements a RESTful service layer for user management, performing creation, retrieval, modification, and deletion operations on a DynamoDB database.

The application automatically validates incoming HTTP request payloads, ensures required attributes are present, and returns standardized responses with CORS-compliant headers. When a user is created without a pre-defined identifier, the system generates a cryptographically secure 16-byte random hexadecimal identifier.

## Features

- **AWS Lambda Integration**: Formatted to handle `events.APIGatewayV2HTTPRequest` inputs and return `events.APIGatewayV2HTTPResponse` outputs using the AWS Lambda Go SDK.
- **DynamoDB CRUD Operations**: Communicates with DynamoDB to retrieve, insert, scan, update, and delete entries via the AWS SDK for Go V2.
- **Request Routing**: Houses an in-app router matching the HTTP method and path for `/users` and `/users/{id}` resource endpoints.
- **CORS Support**: Serves default response headers including `Access-Control-Allow-Origin: *`, `Access-Control-Allow-Methods`, and `Access-Control-Allow-Headers`.
- **Validation**: Enforces existence and formats for mandatory user attributes (`name`, `email`, `username`, `birthdate`).
- **Secure ID Generation**: Utilizes cryptographically secure random bytes to generate unique IDs if not provided during creation.

## Architecture

The project is structured according to a layered architecture pattern:

- **Entry Point (`main.go`)**: Initializes dependencies (AWS DynamoDB client, repository, service, and controller) and starts the AWS Lambda runtime loop using `lambda.Start`.
- **Controller Layer (`internal/controller/user_controller.go`)**: Validates HTTP routes/methods, deserializes incoming JSON data, hands execution to the service layer, and maps service/repository errors to HTTP status codes.
- **Service Layer (`internal/service/user_service.go`)**: Implements business and validation rules (e.g., verifying fields, appending timestamps, generating identifiers).
- **Repository Layer (`internal/repository/`)**: Defines database interactions. `dynamo_user_repository.go` implements the repository interface utilizing DynamoDB API calls (`Scan`, `GetItem`, `PutItem`, `DeleteItem`).
- **Data Model (`internal/model/user.go`)**: Declares the `User` struct layout mapping to both JSON strings and DynamoDB attribute values.

## Technologies Used

| Category | Technology |
| :--- | :--- |
| Programming Language | Go (v1.26.4) |
| Core Framework | AWS Lambda Go (`github.com/aws/aws-lambda-go`) |
| SDK | AWS SDK for Go V2 (`github.com/aws/aws-sdk-go-v2`) |
| Database | Amazon DynamoDB |

## Requirements

- Go version `1.26.4` or newer.
- An AWS DynamoDB table configured with a string-type partition key named `id`.

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

The application is configured using the following environment variable:

| Variable | Required | Description |
| :--- | :--- | :--- |
| `USERS_TABLE` | Yes | The name of the Amazon DynamoDB table where user records are stored. |

## Execution

The compiled binary runs inside the AWS Lambda environment upon receiving events from Amazon API Gateway V2.

To execute the Go process directly (e.g., for local testing tools that invoke binaries), ensure the `USERS_TABLE` environment variable is defined:

```bash
USERS_TABLE=users ./bootstrap
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

### JSON Payloads

#### Create User Request (`POST /users`)
```json
{
  "name": "Jane Doe",
  "email": "jane.doe@example.com",
  "username": "janedoe",
  "birthdate": "1995-05-15"
}
```

#### Successful Response (`201 Created`)
```json
{
  "id": "7f0980cb630dbf3e5898d89e44efb4ca",
  "name": "Jane Doe",
  "email": "jane.doe@example.com",
  "username": "janedoe",
  "birthdate": "1995-05-15",
  "creationdate": "2026-06-19T16:20:00Z"
}
```

## Project Structure

```text
rs-lambda-go/
├── internal/
│   ├── controller/
│   │   └── user_controller.go
│   ├── model/
│   │   └── user.go
│   ├── repository/
│   │   ├── dynamo_user_repository.go
│   │   └── user_repository.go
│   └── service/
│       └── user_service.go
├── .env
├── .gitignore
├── go.mod
├── go.sum
└── main.go
```

## Security

- **Payload Validation**: The service layer validates that the user payload contains non-empty `name`, `email`, `username`, and `birthdate` values, and checks that the `email` string contains the `@` symbol.
- **Database Write Guard**: The database repository uses conditional expressions (`attribute_not_exists(id)`) on `Create` and (`attribute_exists(id)`) on `Delete` to prevent duplicate creations or deletion of non-existent keys.
- **Cross-Origin Resource Sharing (CORS)**: All controllers return HTTP headers allowing headers and methods standard to REST API specifications across any origin.
