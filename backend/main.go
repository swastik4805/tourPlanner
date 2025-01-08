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
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name      string               `bson:"name" json:"name"`
	StartDate string               `bson:"startdate" json:"startdate"`
	CreatedAt time.Time            `bson:"createdat" json:"createdat"`
	userIDs   []primitive.ObjectID `bson:"tour_ids" json:"tour_ids"`
}

type User struct {
	ID      primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name    string               `bson:"name" json:"name"`
	Email   string               `bson:"email" json:"email"`
	TourIDs []primitive.ObjectID `bson:"user_ids" json:"user_ids"`
}

var dbName = "tourPlanner"
var tourCollection *mongo.Collection
var deletedCollection *mongo.Collection
var userCollection *mongo.Collection

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
	userCollection = client.Database(dbName).Collection("users")
	println("collection instance in ready")
}

func CreateTour(w http.ResponseWriter, r *http.Request) {
	var tour Tour
	_ = json.NewDecoder(r.Body).Decode(&tour)
	tour.CreatedAt = time.Now()
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

func AddUserToRoute(w http.ResponseWriter, r *http.Request) {
	//get the tour_id from the url
	params := mux.Vars(r)
	tour_id, _ := primitive.ObjectIDFromHex(params["tour_id"])

	//get the id of the user from the body
	var userID primitive.ObjectID
	err := json.NewDecoder(r.Body).Decode(&userID)
	if err != nil {
		fmt.Println("some problem with the request data")
		return
	}

	//find the tour with the tour_id
	filter := bson.M{"_id": tour_id}
	update := bson.M{"$addToSet": bson.M{"user_ids": userID}}
	updated, err := tourCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println("some error occured while adding in db")
		return
	}
	fmt.Println("added user to tour", updated)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bson.M{"message": "user added successfully"})

}

func removeUserFromTour(w http.ResponseWriter, r *http.Request) {
	//get the tour_id from the url
	params := mux.Vars(r)
	tour_id, err2 := primitive.ObjectIDFromHex(params["tour_id"])
	if err2 != nil {
		fmt.Println("some problem in parsing tour_id")
		return
	}
	//get the id of the user from the body
	var userID primitive.ObjectID
	err := json.NewDecoder(r.Body).Decode(&userID)
	fmt.Println(userID)
	if err != nil {
		fmt.Println("some problem with the request data")
		return
	}
	//find the tour with the tour_id
	filter := bson.M{"_id": tour_id}
	update := bson.M{"$pull": bson.M{"user_ids": userID}}
	deleted, err1 := tourCollection.UpdateOne(context.TODO(), filter, update)
	if err1 != nil {
		fmt.Println("some error occured while removing from db")
		return
	}
	fmt.Println("removed user", deleted)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/createTour", CreateTour).Methods("POST")
	r.HandleFunc("/updateTour/{id}", UpdateTour).Methods("PUT")
	r.HandleFunc("/deleteTour/{id}", deleteTour).Methods("DELETE")
	r.HandleFunc("/addUserToTour/{tour_id}", AddUserToRoute).Methods("POST")
	r.HandleFunc("/removeUserFromTour/{tour_id}", removeUserFromTour).Methods("DELETE")
	http.ListenAndServe(":8080", r)
}
