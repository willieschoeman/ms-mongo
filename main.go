package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo"
)

/* 
-----------------------
-- DB CONFIG SECTION --
-----------------------
*/

var client *mongo.Client

/* 
----------------------
-- HANDLERS SECTION --
----------------------
*/

// Insert a document
func Insert(response http.ResponseWriter, request *http.Request) {
	
	response.Header().Set("content-type", "application/json")

	params := mux.Vars(request)
	db := params["db"]
	coll := params["coll"]

	var document interface{}

	err := json.NewDecoder(request.Body).Decode(&document)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	collection := client.Database(db).Collection(coll)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.InsertOne(ctx, document)

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	json.NewEncoder(response).Encode(result)
}

// Get document(s)
func Get(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("content-type", "application/json")

	params := mux.Vars(request)
	db := params["db"]
	coll := params["coll"]

	query := bson.M{}
	var documents []interface{}

	if len(request.URL.Query()) != 0 {

		queryParams := request.URL.Query()

		for key, value := range queryParams {
			query[key] = value[0]
		}
	}

	collection := client.Database(db).Collection(coll)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, query)

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var document interface{}
		cursor.Decode(&document)
		documents = append(documents, document)
	}

	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	json.NewEncoder(response).Encode(documents)

}

// Update document(s)
func Update(response http.ResponseWriter, request *http.Request) {
	
	response.Header().Set("content-type", "application/json")

	params := mux.Vars(request)
	db := params["db"]
	coll := params["coll"]

	query := bson.M{}
	var document interface{}

	if len(request.URL.Query()) != 0 {

		queryParams := request.URL.Query()

		for key, value := range queryParams {
			query[key] = value[0]
		}
	}

	err := json.NewDecoder(request.Body).Decode(&document)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	collection := client.Database(db).Collection(coll)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.UpdateMany(ctx, query, bson.M{"$set": document})

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	json.NewEncoder(response).Encode(result)
}

// Delete document(s)
func Delete(response http.ResponseWriter, request *http.Request) {
	
	response.Header().Set("content-type", "application/json")

	params := mux.Vars(request)
	db := params["db"]
	coll := params["coll"]

	query := bson.M{}

	if len(request.URL.Query()) != 0 {

		queryParams := request.URL.Query()

		for key, value := range queryParams {
			query[key] = value[0]
		}
	}

	collection := client.Database(db).Collection(coll)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.DeleteMany(ctx, query)

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	json.NewEncoder(response).Encode(result)
}

/* 
------------------
-- MAIN SECTION --
------------------
*/

// Start Of Program
func main() {
	
	var err error

	log.Println("Starting the application...")

	// Set context and options
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	
	// Connect to mongo
	client, err = mongo.Connect(ctx, clientOptions)
	
	if err != nil {
		log.Fatal(err)
	}
	
	// Check the connection
	err = client.Ping(ctx, nil)
	
	if err != nil {
		log.Fatal(err)
	}
	
	// Successfully connected
	log.Println("Connected to DB!")

	// Set allowed handlers
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	// New router and listen
	router := mux.NewRouter()
	router.HandleFunc("/mongo/{db}/{coll}", Insert).Methods("POST")
	router.HandleFunc("/mongo/{db}/{coll}", Get).Methods("GET")
	router.HandleFunc("/mongo/{db}/{coll}", Update).Methods("PUT")
	router.HandleFunc("/mongo/{db}/{coll}", Delete).Methods("DELETE")
	log.Println("Router listening...")
	log.Fatal(http.ListenAndServe(":32345", handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}