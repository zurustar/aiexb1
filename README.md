# Schedule Sharing Web Service

## About

This is a web service for managing and sharing schedules. Users can register, log in, and manage their schedules through a RESTful API.

## Features

*   User registration and login
*   JWT-based authentication
*   Create, Read, Update, and Delete (CRUD) operations for schedules
*   RESTful API

## Requirements

*   Go (version 1.22 or later)

## Build

To build the application, run the following command in the root directory:

```bash
go build -o schedule-app ./cmd/app
```

This will create an executable file named `schedule-app` in the root directory.

## Run

To run the application, you need to set the `JWT_SECRET` environment variable. This secret is used to sign and verify JWTs for authentication.

**For development:**

You can use a simple, temporary secret.

```bash
export JWT_SECRET="your-development-secret-key"
./schedule-app
```

**Using `go run`:**

You can also run the application directly without building it first:

```bash
export JWT_SECRET="your-development-secret-key"
go run ./cmd/app/main.go
```

The server will start on port `8080`.

## API Usage

You can interact with the API using a tool like `curl`.

### Register a new user

To register a new user, send a `POST` request to the `/api/users/register` endpoint:

```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}' http://localhost:8080/api/users/register
```

On success, you will receive a `201 Created` status and the user's information in the response body.

### Log in

To log in, send a `POST` request to the `/api/users/login` endpoint with the user's email and password:

```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "email": "test@example.com",
  "password": "password123"
}' http://localhost:8080/api/users/login
```

On success, you will receive a `200 OK` status and a JWT token in the response body. This token should be included in the `Authorization` header for all subsequent authenticated requests.

Example response:

```json
{
  "token": "your.jwt.token"
}
```