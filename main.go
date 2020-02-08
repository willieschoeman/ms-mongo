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

// Action a document
func Action(response http.ResponseWriter, request *http.Request) {
	
	response.Header().Set("content-type", "application/json")

	// Params
	params := mux.Vars(request)
	db := params["db"]
	coll := params["coll"]

	// Check Params
	if db == "" {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "Missing DB!" }`))
		return
	}

	if coll == "" {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "Missing Collection!" }`))
		return
	}

	// Decode Body
	var document interface{}

	err := json.NewDecoder(request.Body).Decode(&document)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	// Get Action
	var action interface{}

	if val, ok := document.(map[string]interface{})["action"]; ok {
		action = val
	} else {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "Missing Action!" }`))
		return
	}

	// Query and data
	var query interface{}
	var data interface{}
	var hasQuery bool
	var hasData bool

	// Get query
	if val, ok := document.(map[string]interface{})["query"]; ok {
		query = val
		hasQuery = ok
	}
	
	// Get data
	if val, ok := document.(map[string]interface{})["data"]; ok {
		data = val
		hasData = ok
	}

	switch action {

	// Insert a document
	case "insert":

		if !hasData {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "Missing Data!" }`))
			return
		}

		collection := client.Database(db).Collection(coll)
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		result, err := collection.InsertOne(ctx, data)

		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}

		json.NewEncoder(response).Encode(result)
		
	// Get a document
	case "get":

		if !hasQuery {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "Missing Query!" }`))
			return
		}

		var documents []interface{}

		collection := client.Database(db).Collection(coll)
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
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

	// Update a document
	case "upate":
		
		if !hasQuery {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "Missing Query!" }`))
			return
		}

		if !hasData {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "Missing Data!" }`))
			return
		}

		collection := client.Database(db).Collection(coll)
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		result, err := collection.UpdateMany(ctx, query, bson.M{"$set": data})
	
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
	
		json.NewEncoder(response).Encode(result)

	// Delete a document
	case "delete":

		if !hasQuery {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "Missing Query!" }`))
			return
		}

		collection := client.Database(db).Collection(coll)
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		result, err := collection.DeleteMany(ctx, query)
	
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
	
		json.NewEncoder(response).Encode(result)
	
	// Default for wrong action
	default:
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "Unknown Action!" }`))
		return
	}

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
	router.HandleFunc("/ms-mongo/{db}/{coll}", Action).Methods("POST")
	log.Println("Router listening...")
	log.Fatal(http.ListenAndServe(":32345", handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}