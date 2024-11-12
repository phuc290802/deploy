package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Todo struct {
	ID        string     `json:"id" validate:"required"`
	Title     string     `json:"title" validate:"required"`
	Desc      string     `json:"description" validate:"max=500"`
	Done      bool       `json:"done"`
	CreatedAt time.Time  `json:"created_at" validate:"required"`
	DoneAt    *time.Time `json:"done_at,omitempty"`
}

type TodoStore interface {
	GetAllTodoDB(ctx context.Context) ([]Todo, error)
	GetTodoByIdDB(ctx context.Context, id string) (Todo, error)
	CreateTodoDB(ctx context.Context, todo Todo) (Todo, error)
	UpdateTodoDB(ctx context.Context, id string, todo Todo) (Todo, error)
	DeleteTodoByIdDB(ctx context.Context, id string) error
	ChangeStatusDB(ctx context.Context, id string) error
}

type Db struct {
	Conn  *pgxpool.Pool
	mutex sync.Mutex
}

func (db *Db) GetAllTodoDB(ctx context.Context) ([]Todo, error) {
	var todos []Todo
	var wg sync.WaitGroup

	// Thêm ORDER BY vào truy vấn để sắp xếp theo ID
	rows, err := db.Conn.Query(context.Background(), "SELECT * FROM todo ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Desc, &todo.Done, &todo.CreatedAt, &todo.DoneAt)
		if err != nil {
			log.Fatal(err)
		}

		wg.Add(1)
		go func(todo Todo) {
			defer wg.Done()
			db.mutex.Lock()
			todos = append(todos, todo)
			db.mutex.Unlock()
		}(todo)
	}
	wg.Wait()

	return todos, nil
}

func (db *Db) GetTodoByIdDB(ctx context.Context, id string) (Todo, error) {
	var todo Todo
	err := db.Conn.QueryRow(context.Background(), "SELECT * FROM todo WHERE id = $1", id).
		Scan(&todo.ID, &todo.Title, &todo.Desc, &todo.Done, &todo.CreatedAt, &todo.DoneAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return Todo{}, fmt.Errorf("todo not found with ID %s: %w", id, ErrTodoNotFound)
		}
		return Todo{}, fmt.Errorf("failed to retrieve todo: %v", err)
	}

	return todo, nil
}

func (db *Db) CreateTodoDB(ctx context.Context, todo Todo) (Todo, error) {
	todo.ID = uuid.New().String()
	todo.CreatedAt = time.Now()

	if todo.Done {
		now := time.Now()
		todo.DoneAt = &now
	} else {
		todo.DoneAt = nil
	}

	err := db.Conn.QueryRow(ctx,
		"INSERT INTO todo (id, title, description, done, created_at, done_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, title, description, done, created_at, done_at",
		todo.ID, todo.Title, todo.Desc, todo.Done, todo.CreatedAt, todo.DoneAt).
		Scan(&todo.ID, &todo.Title, &todo.Desc, &todo.Done, &todo.CreatedAt, &todo.DoneAt)

	if err != nil {
		return Todo{}, err
	}

	return todo, nil
}

func (db *Db) UpdateTodoDB(ctx context.Context, id string, todo Todo) (Todo, error) {
	existingTodo, err := db.GetTodoByIdDB(ctx, id)
	if err != nil {
		return Todo{}, err
	}
	todo.CreatedAt = existingTodo.CreatedAt

	if todo.Done {
		now := time.Now()
		todo.DoneAt = &now
	} else {
		todo.DoneAt = nil
	}

	var updatedTodo Todo
	err = db.Conn.QueryRow(context.Background(),
		"UPDATE todo SET title=$1, description=$2, done=$3, done_at=$4 WHERE id=$5 RETURNING id, title, description, done, created_at, done_at",
		todo.Title, todo.Desc, todo.Done, todo.DoneAt, id).
		Scan(&updatedTodo.ID, &updatedTodo.Title, &updatedTodo.Desc, &updatedTodo.Done, &updatedTodo.CreatedAt, &updatedTodo.DoneAt)

	if err != nil {
		return Todo{}, err
	}

	return updatedTodo, nil
}

func (db *Db) DeleteTodoByIdDB(ctx context.Context, id string) error {

	_, err := db.GetTodoByIdDB(ctx, id)
	if err != nil {
		return err
	}

	_, err = db.Conn.Exec(context.Background(), "DELETE FROM todo WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %v", err)
	}

	return nil
}

func (db *Db) ChangeStatusDB(ctx context.Context, id string) error {
	var todo Todo

	err := db.Conn.QueryRow(ctx, "SELECT done FROM todo WHERE id = $1", id).Scan(&todo.Done)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("todo not found with ID %s: %w", id, ErrTodoNotFound)
		}
		return fmt.Errorf("failed to retrieve todo: %v", err)
	}

	newDoneStatus := !todo.Done
	var doneAt interface{}
	if newDoneStatus {
		now := time.Now()
		doneAt = now
	} else {
		doneAt = nil
	}

	_, err = db.Conn.Exec(ctx, "UPDATE todo SET done = $1, done_at = $2 WHERE id = $3", newDoneStatus, doneAt, id)
	if err != nil {
		return fmt.Errorf("failed to update todo status: %v", err)
	}

	return nil
}
