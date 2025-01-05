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

var tourCollection *mongo.Collection
var deletedCollection *mongo.Collection

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return
	}
	println("Mongo db connected")
	tourCollection = client.Database(dbName).Collection("tours")
	deletedCollection = client.Database(dbName).Collection("deletedTours")
	println("collection instance in ready")
}

func CreateTour(w http.ResponseWriter, r *http.Request) {
	var tour Tour
	_ = json.NewDecoder(r.Body).Decode(&tour)
	inserted, err := tourCollection.InsertOne(context.Background(), tour)
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
	updated, err := tourCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("updated one tour", updated)
}

func deleteTour(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	filter := bson.M{"_id": id}
	var tour Tour
	err2 := tourCollection.FindOne(context.Background(), filter).Decode(&tour)
	if err2 != nil {
		http.Error(w, "problem in err2", http.StatusBadRequest)
		return
	}
	deleted, err := tourCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		http.Error(w, "problem in err", http.StatusBadRequest)
		return
	}
	fmt.Println("deleted one tour", deleted)
	inserted, err1 := deletedCollection.InsertOne(context.Background(), tour)
	if err1 != nil {
		http.Error(w, "problem in err1", http.StatusBadRequest)
		return
	}
	fmt.Println("inserted one tour", inserted.InsertedID)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/createTour", CreateTour).Methods("POST")
	r.HandleFunc("/updateTour/{id}", UpdateTour).Methods("PUT")
	r.HandleFunc("/deleteTour/{id}", deleteTour).Methods("DELETE")
	http.ListenAndServe(":8080", r)
}
