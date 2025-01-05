package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

type Tour struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	StartDate string             `bson:"startdate" json:"startdate"`
	CreatedAt time.Time          `bson:"createdat" json:"createdat"`
}

var dbName = "tourPlanner"
var colName = "tours"
var collection *mongo.Collection

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return
	}
	println("Mongo db connected")
	collection = client.Database(dbName).Collection(colName)
	println("collection instance in ready")
}

func insertOneTour(tour Tour) {
	inserted, err := collection.InsertOne(context.Background(), tour)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("inserted one tour", inserted.InsertedID)
}

func CreateTour(w http.ResponseWriter, r *http.Request) {
	var tour Tour
	_ = json.NewDecoder(r.Body).Decode(&tour)
	insertOneTour(tour)
	json.NewEncoder(w).Encode(tour)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/createTour", CreateTour).Methods("POST")
	http.ListenAndServe(":8080", r)
}
