package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname  string             `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

var client *mongo.Client

//Agragar a una persona
func CreatePersonEndpoint(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("content-type", "application/json")
	var person Person
	_ = json.NewDecoder(req.Body).Decode(&person)
	colletion := client.Database("thepolyglotdeveloper").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := colletion.InsertOne(ctx, person)
	json.NewEncoder(w).Encode(result)

}

//Lista de Personas
func GetPeopleEndpoint(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("content-type", "application/json")
	var people []Person
	collection := client.Database("thepolyglotdeveloper").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var person Person
		cursor.Decode(&person)
		people = append(people, person)
	}
	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(w).Encode(people)
}

//Ver una persona
func GetPersonEndpoint(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("content-type", "application/json")                         //tipo de llave y valor e la que se escribe desde el cliente
	params := mux.Vars(req)                                                    //Los datos guardados en el servidor
	id, _ := primitive.ObjectIDFromHex(params["id"])                           //Asigna ID hexadecimal del cliente al id. Se desprecia el error
	var person Person                                                          //Crea una variable tipo Person tipo subyasente struct
	collection := client.Database("thepolyglotdeveloper").Collection("people") //Crea una coleccion en base de datos
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)        //Crea un contexto dentro la funcion
	err := collection.FindOne(ctx, Person{ID: id}).Decode(&person)             //Busca a la persona por el id y devuele el documnto que contiene el id y lo codifica en json.
	if err != nil {                                                            //Si hay un error en la decodificacion
		w.WriteHeader(http.StatusInternalServerError)           //Muestra error
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`)) //Escribe en w en formato json un mensaje
		return
	}
	json.NewEncoder(w).Encode(person)
}

func main() {
	fmt.Println("Comenzando Aplicacion...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	router := mux.NewRouter()
	router.HandleFunc("/person", CreatePersonEndpoint).Methods("POST")
	router.HandleFunc("/people", GetPeopleEndpoint).Methods("GET")
	router.HandleFunc("/person/{id}", GetPersonEndpoint).Methods("GET")
	http.ListenAndServe(":12345", router)
	// "mongodb://localhost:27017"
}
