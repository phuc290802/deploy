package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
)

type APIHandler struct {
	todoStore TodoStore
}

var ErrTodoNotFound = errors.New("todo not found")

func NewTodoHandler(todoStore TodoStore) *APIHandler {
	return &APIHandler{todoStore: todoStore}
}

// @Summary Get all todos
// @Description Retrieve all todo items from the database
// @Tags Todos
// @Accept json
// @Produce json
// @Success 200 {array} Todo "OK"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /todos [get]
// Hàm xử lý lỗi trả về JSON hợp lệ
func (h *APIHandler) GetAllTodos(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	todos, err := h.todoStore.GetAllTodoDB(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "no data found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(todos)
}

// @Summary Get a todo by ID
// @Description Retrieve a todo item by its ID from the database
// @Tags Todos
// @Accept json
// @Produce json
// @Param id path string true "Todo ID"
// @Success 200 {object} Todo "OK"
// @Failure 404 {object} ErrorResponse "Todo not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /todo/{id} [get]
func (h *APIHandler) GetTodoByID(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	idStr := vars["id"]

	todo, err := h.todoStore.GetTodoByIdDB(ctx, idStr)
	if err != nil {
		if errors.Is(err, ErrTodoNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "todo not found"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal Server Error"})
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(todo)
}

// @Summary Create a new todo
// @Description Create a new todo item in the database
// @Tags Todos
// @Accept json
// @Produce json
// @Param todo body Todo true "Todo to create"
// @Success 201 {object} Todo "Created"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 500 {object} ErrorResponse "Failed to create todo"
// @Router /todo [post]
func (h *APIHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var todo Todo
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Request body is empty"})
		return
	}

	createdTodo, err := h.todoStore.CreateTodoDB(ctx, todo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create todo: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTodo)
}

// @Summary Update an existing todo
// @Description Update an existing todo item in the database by ID
// @Tags Todos
// @Accept json
// @Produce json
// @Param id path string true "Todo ID"
// @Param todo body Todo true "Updated todo data"
// @Success 200 {object} Todo "Updated"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 404 {object} ErrorResponse "Todo not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /todo/{id} [put]
func (h *APIHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	idStr := vars["id"]

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Request body is empty"})
		return
	}

	updatedTodo, err := h.todoStore.UpdateTodoDB(ctx, idStr, todo)
	if err != nil {
		if errors.Is(err, ErrTodoNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "todo not found"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal Server Error"})
		}
		return
	}

	// Send the updated Todo as the response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTodo)
}

// @Summary Delete a todo by ID
// @Description Delete a todo item from the database by ID
// @Tags Todos
// @Accept json
// @Produce json
// @Param id path string true "Todo ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse "Todo not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /todo/{id} [delete]
func (h *APIHandler) DeleteTodoByID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	idStr := vars["id"]

	err := h.todoStore.DeleteTodoByIdDB(ctx, idStr)
	if err != nil {
		if errors.Is(err, ErrTodoNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Todo not found with ID " + idStr})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete todo: " + err.Error()})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Change the status of a todo
// @Description Change the status of a todo item by its ID
// @Tags Todos
// @Accept json
// @Produce json
// @Param id path string true "Todo ID"
// @Success 200 {object} StatusResponse "Status changed successfully"
// @Failure 404 {object} ErrorResponse "Todo not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /todo/changeStatus/{id} [post]
func (h *APIHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	idStr := vars["id"]

	err := h.todoStore.ChangeStatusDB(ctx, idStr)
	if err != nil {
		if errors.Is(err, ErrTodoNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Todo not found"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to changeStatus todo: " + err.Error()})
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}
