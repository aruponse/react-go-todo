package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title string             `json:"title" bson:"title"`
	Done  bool               `json:"done" bson:"done"`
}

var collection *mongo.Collection

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	DB_URI := os.Getenv("DB_URI")
	client_options := options.Client().ApplyURI(DB_URI)
	client, err := mongo.Connect(context.Background(), client_options)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	collection = client.Database("react_go_todos").Collection("todos")

	app := fiber.New()

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Delete("/api/todos/:id", deleteTodo)
	app.Put("/api/todos/:id", updateTodo)
	app.Patch("/api/todos/:id", toggleTodo)

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "4000"
	}

	log.Fatal(app.Listen("0.0.0.0:" + PORT))
}

func getTodos(c *fiber.Ctx) error {
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var todos []Todo
	err = cursor.All(context.Background(), &todos)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(todos)
}

func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)

	if err := c.BodyParser(todo); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	result, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	todo.ID = result.InsertedID.(primitive.ObjectID)

	return c.JSON(todo)
}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(204)
}

func updateTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	todo := new(Todo)

	if err := c.BodyParser(todo); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$set": todo})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(todo)
}

func toggleTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	todo := new(Todo)

	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(todo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	todo.Done = !todo.Done

	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$set": bson.M{"done": todo.Done}})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(todo)
}
