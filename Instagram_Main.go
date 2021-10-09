package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type User struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name         string             `json:"name,omitempty" bson:"name,omitempty"`
	Password     string             `json:"password,omitempty" bson:"-"`
	Email        string             `json:"email,omitempty" bson:"email,omitempty"`
	PasswordHash string             `json:"passwordhash,omitempty" bson:"passwordhash,omitempty"`
}

type Post struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserName        string             `json:"username,omitempty" bson:"username,omitempty"`
	Password        string             `json:"password,omitempty" bson:"-"`
	Caption         string             `json:"caption" bson:"caption"`
	ImageURL        string             `json:"imageurl,omitempty" bson:"imageurl,omitempty"`
	PostedTimeStamp string             `json:"postedtimestamp,omitempty" bson:"postedtimestamp,omitempty"`
}

func RetrieveListOfPost(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var user User
	var AllPosts []Post
	UserCollection := client.Database("InstagramAPI").Collection("AllUsers")
	user_ptr, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := UserCollection.FindOne(user_ptr, User{ID: id}).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	if id == user.ID {
		PostCollection := client.Database("InstagramAPI").Collection("AllPosts")
		post_ptr, _ := context.WithTimeout(context.Background(), 30*time.Second)
		cursor, err := PostCollection.Find(post_ptr, Post{UserName: user.Name})
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		defer cursor.Close(post_ptr)
		for cursor.Next(post_ptr) {
			fmt.Println("For loop entered")
			var post Post
			cursor.Decode(&post)
			AllPosts = append(AllPosts, post)
		}
		if err := cursor.Err(); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		json.NewEncoder(response).Encode(AllPosts)
	} else {
		fmt.Println("User Not Found")
		return
	}
}

func main() {
	fmt.Println("Initialized")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/users", User_Create).Methods("POST")
	router.HandleFunc("/users/{id}", RetreiveUser).Methods("GET")
	router.HandleFunc("/posts", Post_Create).Methods("POST")
	router.HandleFunc("/posts/{id}", RetreivePost).Methods("GET")
	router.HandleFunc("/posts/users/{id}", RetrieveListOfPost).Methods("GET")
	router.HandleFunc("/AllUsers", GetAllUsers).Methods("GET")
	router.HandleFunc("/AllPosts", GetAllPosts).Methods("GET")
	http.ListenAndServe(":8888", router)
}
