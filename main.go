package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"net/http"
	"time"
)

var db *gorm.DB

type Todo struct {
	gorm.Model
	Todo  string
	Done bool
}

func hello(w http.ResponseWriter, req *http.Request) {

	now := time.Now()

	fmt.Fprintf(w, "%v", now)
}

func headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func getAllTodos (w http.ResponseWriter, req *http.Request) {

	var allTodos []Todo
	db.Find(&allTodos)


	//fmt.Fprintf(w, Todo[])
}

func main() {

	var err error
	db, err = gorm.Open(postgres.Open("postgresql://localhost:5432/woopy"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	err = db.AutoMigrate(&Todo{})
	if err != nil {
		panic("failed to migrate table")
	}

	r := mux.NewRouter()
	r.HandleFunc("/", hello)
	r.HandleFunc("/todo", getAllTodos).Methods("GET")
	r.HandleFunc("/todo", createTodo).Methods("POST")
	http.Handle("/", r)

	fmt.Println("Starting server on localhost:8090")
	err = http.ListenAndServe(":8090", nil)
	if err != nil {
		fmt.Println("Oh nooo!")
		fmt.Print(err)
	}
}

type TodoJson struct {
	Todo string
}

type ErrorResponse struct {
	Error string
}
// createTodo is our method of creating todos
func createTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	newTodo := TodoJson{}

	input, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Failed to read body")
		errorRes := ErrorResponse{
			Error: err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		response, err := json.Marshal(&errorRes)
		if err != nil {
			fmt.Println(err)
		}
		_, err = w.Write(response)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(input, &newTodo)
	if err != nil {
		fmt.Println(err)
		errorRes := ErrorResponse{
			Error: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		response, err := json.Marshal(&errorRes)
		if err != nil {
			fmt.Println(err)
		}
		_, err = w.Write(response)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	dbTodo := Todo{
		Todo: newTodo.Todo,
		Done: false,
	}
	db.Create(&dbTodo)
	w.WriteHeader(http.StatusCreated)
	response, err := json.Marshal(&dbTodo)
	if err != nil {
		fmt.Println(err)
	}
	_, err = w.Write(response)
	if err != nil {
		fmt.Println(err)
	}
}