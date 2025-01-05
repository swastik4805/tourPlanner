package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func CreateTour(w http.ResponseWriter, r *http.Request) {
	var tour Tour
	_ = json.NewDecoder(r.Body).Decode(&tour)
	inserted, err := collection.InsertOne(context.Background(), tour)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("inserted one tour", inserted.InsertedID)
	json.NewEncoder(w).Encode(tour)
}

func UpdateTour(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var tour Tour
	_ = json.NewDecoder(r.Body).Decode(&tour)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": tour}
	updated, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("updated one tour", updated)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/createTour", CreateTour).Methods("POST")
	r.HandleFunc("/updateTour/{id}", UpdateTour).Methods("PUT")
	http.ListenAndServe(":8080", r)
}
