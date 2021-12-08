package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"net/http"
)

var db *gorm.DB

type Todo struct {
	gorm.Model
	Todo string
	Done bool
}

func home(w http.ResponseWriter, req *http.Request) {
	body := `<body>
	<ul>
		<li>GET <a href="/todo">/todo</a></li>
		<li>POST <a href="/todo">/todo</a></li>
	</ul>
</body`

	fmt.Fprintf(w, "%v", body)
}

func headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func getAllTodos(w http.ResponseWriter, req *http.Request) {
	var allTodos []Todo
	db.Find(&allTodos)

	myTodos, err := json.Marshal(&allTodos)
	if err != nil {
		handleErr(err, w, http.StatusInternalServerError)
		return
	}


	w.Write(myTodos)
}

func applicationSetupMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.Header().Set("server", "Toaster,V1.3;")
		next.ServeHTTP(w, r)
    })
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
	r.HandleFunc("/", home)
	r.HandleFunc("/todo", getAllTodos).Methods("GET")
	r.HandleFunc("/todo", createTodo).Methods("POST")
	r.Use(applicationSetupMiddleware)
	http.Handle("/", r)

	fmt.Println("Starting server on http://localhost:8090")
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

func handleErr(err error, w http.ResponseWriter, responseCode int) {
	errorRes := ErrorResponse{
		Error: err.Error(),
	}
	w.WriteHeader(responseCode)
	response, err := json.Marshal(&errorRes)
	if err != nil {
		fmt.Println(err)
	}
	_, err = w.Write(response)
	if err != nil {
		fmt.Println(err)
	}
}

// createTodo is our method of creating todos
func createTodo(w http.ResponseWriter, r *http.Request) {

	newTodo := TodoJson{}

	input, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Failed to read body")
		handleErr(err, w, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(input, &newTodo)
	if err != nil {
		handleErr(err, w, http.StatusBadRequest)
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
