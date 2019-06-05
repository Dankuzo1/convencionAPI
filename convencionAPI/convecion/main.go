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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Person struct {
	CI        string     `json:"ci,omitempty" bson:"ci,omitempty"`
	Edad      int        `json:"edad,omitempty" bson:"edad,omitempty"`
	Nombre    string     `json:"nombre,omitempty" bson:"nombre,omitempty"`
	Apellido  string     `json:"apellido,omitempty" bson:"apellido,omitempty"`
	Direccion *Direccion `json:"direccion,omitempty" bson:"direccion,omitempty"`
}

type Direccion struct {
	Ciudad string `json:"ciudad,omitempty" bson:"ciudad,omitempty"`
	Estado string `json:"estado,omitempty" bson:"estado,omitempty"`
}

var client *mongo.Client

// EndPoints

func CreatePersonEndpoint(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("content-type", "application/json")
	var person Person
	_ = json.NewDecoder(req.Body).Decode(&person)
	colletion := client.Database("Convencion").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := colletion.InsertOne(ctx, person)
	json.NewEncoder(w).Encode(result)

}

func GetPeopleEndpoint(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("content-type", "application/json")
	var people []Person
	collection := client.Database("Convencion").Collection("people")
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

func GetPersonEndpoint(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("content-type", "application/json")                  //tipo de llave y valor e la que se escribe desde el cliente
	params := mux.Vars(req)                                             //Los datos guardados en el servidor
	ci := params["ci"]                                                  //Asigna ci del cliente al programa
	var person Person                                                   //Crea una variable tipo Person tipo subyasente struct
	collection := client.Database("Convencion").Collection("people")    //Crea una coleccion en base de datos
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second) //Crea un contexto dentro la funcion
	err := collection.FindOne(ctx, Person{CI: ci}).Decode(&person)      //Busca a la persona por el id y devuele el documnto que contiene el id y lo codifica en json.
	if err != nil {                                                     //Si hay un error en la decodificacion
		w.WriteHeader(http.StatusInternalServerError)           //Muestra error
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`)) //Escribe en w en formato json un mensaje
		return
	}
	json.NewEncoder(w).Encode(person)
}

func DeletePersonEndpoint(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("content-type", "application/json") //tipo de llave y valor e la que se escribe desde el cliente
	params := mux.Vars(req)                            //Los datos guardados en el servidor
	ci := params["ci"]
	collection := client.Database("Convencion").Collection("people")    //Crea una coleccion en base de datos
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second) //Crea un contexto dentro la funcion
	_, err := collection.DeleteOne(ctx, bson.M{"ci": ci})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)           //Muestra error
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`)) //Escribe en w en formato json un mensaje
		return
	}
	json.NewEncoder(w).Encode(&Person{})
}

func main() {
	fmt.Println("Ejecuntando Aplicacion...")
	router := mux.NewRouter()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	//Endpoints
	router.HandleFunc("/people", GetPeopleEndpoint).Methods("GET")
	router.HandleFunc("/people/{ci}", GetPersonEndpoint).Methods("GET")
	router.HandleFunc("/people", CreatePersonEndpoint).Methods("POST")
	router.HandleFunc("/people/{ci}", DeletePersonEndpoint).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":3000", router))
}
