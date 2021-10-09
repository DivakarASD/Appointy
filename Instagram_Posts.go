package main

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Post_Create(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	user := new(User)
	post := new(Post)
	_ = json.NewDecoder(request.Body).Decode(&post)
	post.PostedTimeStamp = time.Now().Format(time.RFC850)
	h := sha1.New()
	h.Write([]byte(post.Password))
	var Pass_Hash string = base64.URLEncoding.EncodeToString(h.Sum(nil))
	UserCollection := client.Database("InstagramAPI").Collection("AllUsers")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := UserCollection.FindOne(ctx, User{Name: post.UserName, PasswordHash: Pass_Hash}).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	if post.UserName == user.Name && Pass_Hash == user.PasswordHash {
		PostCollection := client.Database("InstagramAPI").Collection("AllPosts")
		result, _ := PostCollection.InsertOne(ctx, post)
		json.NewEncoder(response).Encode(result)
	} else {
		fmt.Println("Invalid UserName or Password")
		return
	}
}

func RetreivePost(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var post Post
	collection := client.Database("InstagramAPI").Collection("AllPosts")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, User{ID: id}).Decode(&post)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(post)
}

func GetAllPosts(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var AllPosts []Post
	collection := client.Database("InstagramAPI").Collection("AllPosts")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
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
}
