package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thedevsaddam/renderer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var rnd *renderer.Render
var collection *mongo.Collection

const (
	mongoURI       = "mongodb://localhost:27017"
	dbName         = "demo_todo"
	collectionName = "todo"
	port           = ":9000"
)

type Todo struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Completed bool               `bson:"completed" json:"completed"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

func init() {
	rnd = renderer.New()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database(dbName).Collection(collectionName)
	log.Println("Connected to MongoDB!")
}

func fetchTodos(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var todos []Todo

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to fetch todo",
			"error":   err.Error(),
		})
		return
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &todos); err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to decode todos",
			"error":   err.Error(),
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todos,
	})
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		rnd.JSON(w, http.StatusBadRequest, err)
		return
	}

	if todo.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title is required",
		})
		return
	}

	todo.ID = primitive.NewObjectID()
	todo.CreatedAt = time.Now()
	todo.Completed = false

	_, err := collection.InsertOne(ctx, todo)
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to save todo",
			"error":   err.Error(),
		})
		return
	}

	rnd.JSON(w, http.StatusCreated, renderer.M{
		"message": "Todo created successfully",
		"todo_id": todo.ID.Hex(),
	})
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id := strings.TrimSpace(chi.URLParam(r, "id"))
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "Invalid ID",
		})
		return
	}

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		rnd.JSON(w, http.StatusBadRequest, err)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"title":     todo.Title,
			"completed": todo.Completed,
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to update todo",
			"error":   err.Error(),
		})
		return
	}

	if result.MatchedCount == 0 {
		rnd.JSON(w, http.StatusNotFound, renderer.M{
			"message": "Todo not found",
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo updated successfully",
	})
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id := strings.TrimSpace(chi.URLParam(r, "id"))
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "Invalid ID",
		})
		return
	}

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to delete todo",
			"error":   err.Error(),
		})
		return
	}

	if result.DeletedCount == 0 {
		rnd.JSON(w, http.StatusNotFound, renderer.M{
			"message": "Todo not found",
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo deleted successfully",
	})
}

func main() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler)
	r.Mount("/todo", todoHandlers())

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-stopChan
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v\n", err)
	}

	log.Println("Server gracefully stopped")
}

func todoHandlers() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", fetchTodos)
		r.Post("/", createTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})
	return rg
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := rnd.Template(w, http.StatusOK, []string{"static/home.tpl"}, nil)
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
