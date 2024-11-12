package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "api/docs"
)

// @title Todo API
// @version 1.0
// @description Đây là API đơn giản để quản lý các công việc.
// @host localhost:8080
// @BasePath /
func main() {
	db, _ := NewDb()
	defer db.Conn.Close()

	h := NewTodoHandler(db)
	router := mux.NewRouter()

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	router.HandleFunc("/swagger/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	router.HandleFunc("/todos", h.GetAllTodos).Methods("GET")
	router.HandleFunc("/todo/{id}", h.GetTodoByID).Methods("GET")
	router.HandleFunc("/todo", h.CreateTodo).Methods("POST")
	router.HandleFunc("/todo/{id}", h.UpdateTodo).Methods("PUT")
	router.HandleFunc("/todo/{id}", h.DeleteTodoByID).Methods("DELETE")
	router.HandleFunc("/todo/changeStatus/{id}", h.ChangeStatus).Methods("POST")

	log.Println("Server đang chạy trên cổng 8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Không thể khởi động server: %v", err)
	}
}
