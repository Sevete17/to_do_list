----------------------------------------------------- ToDo Application -------------------------------------------------------

A simple RESTful API service for managing todo items, built with Go and MongoDB.
Prerequisites
Before running this application, make sure you have the following installed:

1. Go (latest version)
2. MongoDB (running on localhost:27017)
3. The following Go packages:
   - github.com/go-chi/chi
   - github.com/go-chi/chi/middleware
   - github.com/thedevsaddam/renderer
   - gopkg.in/mgo.v2

Installation!!!

Clone the repository
Install dependencies:
  - go mod init todo-app
  - go get github.com/go-chi/chi
  - go get github.com/go-chi/chi/middleware
  - go get github.com/thedevsaddam/renderer
  - go get gopkg.in/mgo.v2

Configuration
The application uses the following default configurations:
  - MongoDB host: localhost:27017
  - Database name: demo_todo
  - Collection name: todo
  - Server port: :9000

Running the Application

Start MongoDB server
Run the application:
  - go run main.go

The server will start on http://localhost:9000

Features

RESTful API design
MongoDB integration
Graceful server shutdown
Request logging middleware
JSON response rendering
Error handling

Error Handling
The API returns appropriate HTTP status codes and error messages:

200: Success
201: Created successfully
400: Bad request (invalid input)
422: Processing error

Shutdown
The application handles graceful shutdown when receiving interrupt signals (Ctrl+C).
