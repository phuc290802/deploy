package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTodoStore struct {
	mock.Mock
}

func (m *MockTodoStore) GetAllTodoDB(ctx context.Context) ([]Todo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, fmt.Errorf("no data found")
	}
	return args.Get(0).([]Todo), args.Error(1)
}

func (m *MockTodoStore) GetTodoByIdDB(ctx context.Context, id string) (Todo, error) {
	args := m.Called(id)
	return args.Get(0).(Todo), args.Error(1)
}

func (m *MockTodoStore) CreateTodoDB(ctx context.Context, todo Todo) (Todo, error) {
	args := m.Called(todo)
	return args.Get(0).(Todo), args.Error(1)
}

func (m *MockTodoStore) UpdateTodoDB(ctx context.Context, id string, todo Todo) (Todo, error) {
	args := m.Called(id, todo)
	return args.Get(0).(Todo), args.Error(1)
}

func (m *MockTodoStore) DeleteTodoByIdDB(ctx context.Context, id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTodoStore) ChangeStatusDB(ctx context.Context, id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTodoStore) Connect(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	args := m.Called(ctx, connStr)
	return nil, args.Error(1)
}

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()

	router.HandleFunc("/todos", h.GetAllTodos).Methods("GET")
	router.HandleFunc("/todo/{id}", h.GetTodoByID).Methods("GET")
	router.HandleFunc("/todo", h.CreateTodo).Methods("POST")
	router.HandleFunc("/todo/{id}", h.UpdateTodo).Methods("PUT")
	router.HandleFunc("/todo/{id}", h.DeleteTodoByID).Methods("DELETE")
	router.HandleFunc("/todo/changeStatus/{id}", h.ChangeStatus).Methods("POST")
	router.ServeHTTP(w, r)
}

///////////// Test success  /////////////

func TestGetAllTodos(t *testing.T) {
	mockStore := new(MockTodoStore)
	mockStore.On("GetAllTodoDB").Return([]Todo{
		{ID: "1", Title: "Todo 1", Desc: "Description 1", Done: false},
		{ID: "2", Title: "Todo 2", Desc: "Description 2", Done: true},
	}, nil)

	handler := NewTodoHandler(mockStore)
	req, _ := http.NewRequest("GET", "/todos", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")
	mockStore.AssertExpectations(t)
}

func TestGetTodoByID(t *testing.T) {
	mockStore := new(MockTodoStore)
	mockStore.On("GetTodoByIdDB", "1").Return(Todo{ID: "1", Title: "Todo 1", Desc: "Description 1", Done: false}, nil)

	handler := NewTodoHandler(mockStore)
	req, _ := http.NewRequest("GET", "/todo/1", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")
	mockStore.AssertExpectations(t)
}

func TestCreateTodo(t *testing.T) {
	mockStore := new(MockTodoStore)
	newTodo := Todo{ID: "3", Title: "Todo 3", Desc: "Description 3", Done: false}
	mockStore.On("CreateTodoDB", newTodo).Return(newTodo, nil)

	handler := NewTodoHandler(mockStore)
	reqBody, _ := json.Marshal(newTodo)
	req, _ := http.NewRequest("POST", "/todo", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Expected status code 201")
	mockStore.AssertExpectations(t)
}

func TestUpdateTodo(t *testing.T) {
	mockStore := new(MockTodoStore)
	updatedTodo := Todo{ID: "1", Title: "Updated Todo", Desc: "Updated Description", Done: true}
	mockStore.On("UpdateTodoDB", "1", updatedTodo).Return(updatedTodo, nil)

	handler := NewTodoHandler(mockStore)
	reqBody, _ := json.Marshal(updatedTodo)
	req, _ := http.NewRequest("PUT", "/todo/1", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")
	mockStore.AssertExpectations(t)
}

func TestDeleteTodoByID(t *testing.T) {
	mockStore := new(MockTodoStore)
	mockStore.On("DeleteTodoByIdDB", "1").Return(nil)

	handler := NewTodoHandler(mockStore)
	req, _ := http.NewRequest("DELETE", "/todo/1", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code, "Expected status code 204")
	mockStore.AssertExpectations(t)
}

func TestChangeStatusTodo(t *testing.T) {
	mockStore := new(MockTodoStore)
	// Mock lại phương thức ChangeStatusDB với đối số là "1"
	mockStore.On("ChangeStatusDB", "1").Return(nil)

	handler := NewTodoHandler(mockStore)

	req, _ := http.NewRequest("POST", "/todo/changeStatus/1", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "expected status code 200")

	mockStore.AssertExpectations(t)
}

///////////// Test Not found ID  /////////////

func TestGetTodoByID_NotFound(t *testing.T) {
	mockStore := new(MockTodoStore)
	mockStore.On("GetTodoByIdDB", "999").Return(Todo{}, ErrTodoNotFound)

	handler := NewTodoHandler(mockStore)
	req, _ := http.NewRequest("GET", "/todo/999", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "Expected status code 404")
	mockStore.AssertExpectations(t)
}

func TestUpdateTodo_NotFound(t *testing.T) {
	mockStore := new(MockTodoStore)
	mockStore.On("UpdateTodoDB", "999", mock.Anything).Return(Todo{}, ErrTodoNotFound)

	handler := NewTodoHandler(mockStore)

	todoJSON := `{"title": "Updated Title", "desc": "Updated Description", "done": false}`
	req, _ := http.NewRequest("PUT", "/todo/999", strings.NewReader(todoJSON))
	req = mux.SetURLVars(req, map[string]string{"id": "999"})

	rr := httptest.NewRecorder()

	handler.UpdateTodo(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "Expected status code 404")
	mockStore.AssertExpectations(t)
}

func TestDeleteTodo_NotFound(t *testing.T) {
	mockStore := new(MockTodoStore)

	mockStore.On("DeleteTodoByIdDB", "999").Return(ErrTodoNotFound)

	handler := NewTodoHandler(mockStore)

	req, _ := http.NewRequest("DELETE", "/todo/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})

	rr := httptest.NewRecorder()

	handler.DeleteTodoByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "Expected status code 404")

	mockStore.AssertExpectations(t)
}

func TestChangeStatus_NotFound(t *testing.T) {
	mockStore := new(MockTodoStore)
	mockStore.On("ChangeStatusDB", "1").Return(ErrTodoNotFound)

	handler := NewTodoHandler(mockStore)

	req, _ := http.NewRequest("POST", "/todo/changeStatus/1", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "Expected tatus code 404")

	mockStore.AssertExpectations(t)
}

///////////// Test Error Db  /////////////

func TestCreateTodo_DBError(t *testing.T) {
	mockStore := new(MockTodoStore)
	newTodo := Todo{ID: "3", Title: "Todo 3", Desc: "Description 3", Done: false}

	mockStore.On("CreateTodoDB", newTodo).Return(Todo{}, errors.New("Database error"))

	handler := NewTodoHandler(mockStore)

	reqBody, _ := json.Marshal(newTodo)
	req, _ := http.NewRequest("POST", "/todo", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500 due to error")

	mockStore.AssertExpectations(t)
}

func TestGetAllTodos_DBError(t *testing.T) {
	mockStore := new(MockTodoStore)
	mockStore.On("GetAllTodoDB").Return(nil, fmt.Errorf("Database connection failed"))

	handler := NewTodoHandler(mockStore)
	req, _ := http.NewRequest("GET", "/todos", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")

	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "no data found", response["error"], "Expected error message to be 'no data found'")

	mockStore.AssertExpectations(t)
}

func TestGetTodoByID_DBError(t *testing.T) {
	mockStore := new(MockTodoStore)

	mockStore.On("GetTodoByIdDB", "1").Return(Todo{}, fmt.Errorf("Database connection failed"))

	handler := NewTodoHandler(mockStore)
	req, _ := http.NewRequest("GET", "/todo/1", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")

	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal Server Error", response["error"], "Expected error message to be 'Internal Server Error'")

	mockStore.AssertExpectations(t)
}

func TestUpdateTodo_DBError(t *testing.T) {
	mockStore := new(MockTodoStore)

	mockStore.On("UpdateTodoDB", "999", mock.Anything).Return(Todo{}, fmt.Errorf("Database connection failed"))

	handler := NewTodoHandler(mockStore)

	req, _ := http.NewRequest("PUT", "/todo/999", bytes.NewBuffer([]byte("{}")))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500 due to error")

	expectedErrorMessage := "Internal Server Error"
	if !strings.Contains(rr.Body.String(), expectedErrorMessage) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrorMessage, rr.Body.String())
	}
	mockStore.AssertExpectations(t)
}

func TestDelete_DBError(t *testing.T) {
	mockStore := new(MockTodoStore)

	mockStore.On("DeleteTodoByIdDB", mock.Anything).Return(fmt.Errorf("Database connection failed"))

	handler := NewTodoHandler(mockStore)

	req, _ := http.NewRequest("DELETE", "/todo/1", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")

	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err, "Expected no error decoding the response")

	assert.Equal(t, "Failed to delete todo: Database connection failed", response["error"], "Expected error message to be 'Failed to delete todo: Database connection failed'")

	mockStore.AssertExpectations(t)
}
func TestChangeStatus_DBError(t *testing.T) {
	mockStore := new(MockTodoStore)
	mockStore.On("ChangeStatusDB", "1").Return(fmt.Errorf("Database connection failed"))

	handler := NewTodoHandler(mockStore)

	req, _ := http.NewRequest("POST", "/todo/changeStatus/1", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")

	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err, "Expected no error decoding the response")

	assert.Equal(t, "Failed to changeStatus todo: Database connection failed", response["error"], "Expected error message to be 'Failed to changeStatus todo: Database connection failed'")

	mockStore.AssertExpectations(t)
}

////////////  Test Error Json Body ////////////////

func TestCreate_JBodyError(t *testing.T) {
	mockStore := new(MockTodoStore)

	mockStore.On("CreateTodoDB", mock.Anything).Return(Todo{}, fmt.Errorf("Request body is empty"))

	handler := NewTodoHandler(mockStore)

	req, _ := http.NewRequest("POST", "/todo", strings.NewReader("invalid json"))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected status code 400")

	expectedErrorMessage := "Request body is empty"
	if !strings.Contains(rr.Body.String(), expectedErrorMessage) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrorMessage, rr.Body.String())
	}
}
func TestUpdateTodo_JBodyError(t *testing.T) {
	mockStore := new(MockTodoStore)

	mockStore.On("UpdateTodoDB", "999", mock.Anything).Return(Todo{}, "Request body is empty")

	handler := NewTodoHandler(mockStore)

	req, _ := http.NewRequest("PUT", "/todo/999", strings.NewReader("invalid json"))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected status code 400")

	expectedErrorMessage := "Request body is empty"
	if !strings.Contains(rr.Body.String(), expectedErrorMessage) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrorMessage, rr.Body.String())
	}
}

//////////// Test connect Db /////////

func TestNewDb_Success(t *testing.T) {
	err := godotenv.Load(".env")
	assert.NoError(t, err, "Error loading .env file")

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	assert.NotEmpty(t, user, "DB_USER environment variable must be set")
	assert.NotEmpty(t, password, "DB_PASSWORD environment variable must be set")
	assert.NotEmpty(t, host, "DB_HOST environment variable must be set")
	assert.NotEmpty(t, port, "DB_PORT environment variable must be set")
	assert.NotEmpty(t, dbname, "DB_NAME environment variable must be set")

	db, _ := NewDb()

	assert.NotNil(t, db, "Expected Db object to be created")
	assert.NotNil(t, db.Conn, "Expected the database connection to be initialized")

	err = db.Conn.Ping(context.Background())
	assert.NoError(t, err, "Expected to successfully ping the database connection")

	defer os.Unsetenv("DB_USER")
	defer os.Unsetenv("DB_PASSWORD")
	defer os.Unsetenv("DB_HOST")
	defer os.Unsetenv("DB_PORT")
	defer os.Unsetenv("DB_NAME")
}

func TestNewDb_FailToLoadEnv(t *testing.T) {
	err := os.Remove(".env")
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove .env file: %v", err)
	}

	_, err = NewDb()
	if err == nil {
		t.Errorf("Expected error loading .env file, but got nil")
	}

	if err != nil && err.Error() != "Error loading .env file: open .env: The system cannot find the file specified." {
		t.Errorf("Expected error 'Error loading .env file' but got: %v", err)
	}

	envContent := []byte(
		"DB_USER=dev\n" +
			"DB_PASSWORD=ujajaQWXa678s6QyvYhc8w\n" +
			"DB_HOST=dev2208-3021.8nk.gcp-asia-southeast1.cockroachlabs.cloud\n" +
			"DB_PORT=26257\n" +
			"DB_NAME=defaultdb\n" +
			"SSL_MODE=verify-full\n")
	err = ioutil.WriteFile(".env", envContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
}

// func TestNewDb_ConnectionError(t *testing.T) {
// 	originalEnvContent, err := ioutil.ReadFile(".env")
// 	if err != nil {
// 		t.Fatalf("Failed to read .env file: %v", err)
// 	}

// 	invalidEnvContent := []byte("DB_USER=invalid_user\nDB_PASSWORD=invalid_password\nDB_HOST=invalid_host\nDB_PORT=5432\nDB_NAME=invalid_db\n")
// 	err = ioutil.WriteFile(".env", invalidEnvContent, 0644)
// 	if err != nil {
// 		t.Fatalf("Failed to write to .env file: %v", err)
// 	}

// 	db, err := NewDb()

// 	if db != nil {
// 		t.Fatalf("Expected nil DB, got: %v", db)
// 	}
// 	if err == nil {
// 		t.Fatalf("Expected error, got nil")
// 	}

// 	expectedError := "Error connecting to the database: failed to connect to `host=invalid_host user=invalid_user database=invalid_db`: hostname resolving error (lookup invalid_host: no such host)"
// 	if err.Error() != expectedError {
// 		t.Fatalf("Expected error message for failed connection, got: %v", err)
// 	}

// 	err = ioutil.WriteFile(".env", originalEnvContent, 0644)
// 	if err != nil {
// 		t.Fatalf("Failed to restore .env file: %v", err)
// 	}

// 	restoredContent, err := ioutil.ReadFile(".env")
// 	if err != nil {
// 		t.Fatalf("Failed to read .env file after restoration: %v", err)
// 	}

// 	if string(restoredContent) != string(originalEnvContent) {
// 		t.Fatalf("The .env file was not restored correctly. Expected: %v, got: %v", originalEnvContent, restoredContent)
// 	}
// }

/////////////End Test Connect Db////////////////

func TestTodoDB(t *testing.T) {
	var ctx context.Context
	db, err := NewDb()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Conn.Close()

	// Case 1: Kiểm tra nhiều mục Todo
	t.Run("CreateTodos", func(t *testing.T) {
		todo_1 := Todo{
			Title:     "Test Todo 1",
			Desc:      "Test Description 1",
			Done:      false,
			CreatedAt: time.Now(),
			DoneAt:    nil,
		}
		todo_2 := Todo{
			Title:     "Test Todo 2",
			Desc:      "Test Description 2",
			Done:      false,
			CreatedAt: time.Now(),
			DoneAt:    nil,
		}

		todo1, err := db.CreateTodoDB(ctx, todo_1)
		if err != nil {
			t.Fatalf("Failed to create todo_1: %v", err)
		}
		todo2, err := db.CreateTodoDB(ctx, todo_2)
		if err != nil {
			t.Fatalf("Failed to create todo_2: %v", err)
		}

		assert.Equal(t, todo1.Title, todo_1.Title)
		assert.Equal(t, todo1.Desc, todo_1.Desc)
		assert.Equal(t, todo1.Done, todo_1.Done)

		assert.Equal(t, todo2.Title, todo_2.Title)
		assert.Equal(t, todo2.Desc, todo_2.Desc)
		assert.Equal(t, todo2.Done, todo_2.Done)
	})

	// Case 2: Lấy tất cả Todos
	t.Run("GetAllTodos", func(t *testing.T) {
		todos, err := db.GetAllTodoDB(ctx)
		if err != nil {
			t.Fatalf("Failed to get all todos: %v", err)
		}

		var foundTodo1, foundTodo2 bool
		for _, todo := range todos {
			if todo.Title == "Test Todo 1" {
				assert.Equal(t, todo.Desc, "Test Description 1")
				assert.Equal(t, todo.Done, false)
				foundTodo1 = true
			}
			if todo.Title == "Test Todo 2" {
				assert.Equal(t, todo.Desc, "Test Description 2")
				assert.Equal(t, todo.Done, false)
				foundTodo2 = true
			}
		}

		assert.True(t, foundTodo1, "Todo 1 was not found in the retrieved list")
		assert.True(t, foundTodo2, "Todo 2 was not found in the retrieved list")
	})

	// Case 3: Get todo By ID
	t.Run("GetTodoByID", func(t *testing.T) {
		todo1, err := db.CreateTodoDB(ctx, Todo{
			Title:     "Test Todo 1",
			Desc:      "Test Description 1",
			Done:      false,
			CreatedAt: time.Now(),
			DoneAt:    nil,
		})
		todoGetByID, err := db.GetTodoByIdDB(ctx, todo1.ID)
		if err != nil {
			t.Fatalf("Failed to get all todos: %v", err)
		}

		assert.Equal(t, todoGetByID.ID, todo1.ID)
	})

	// Case 4: Update Todo
	t.Run("UpdateTodoByID", func(t *testing.T) {
		todo1, err := db.CreateTodoDB(ctx, Todo{
			Title:     "Test Todo 1",
			Desc:      "Test Description 1",
			Done:      false,
			CreatedAt: time.Now(),
			DoneAt:    nil,
		})
		todoUpdate, err := db.UpdateTodoDB(ctx, todo1.ID, Todo{
			Title: "Test Todo 2",
			Desc:  "Test Description 2",
		})
		if err != nil {
			t.Fatalf("Failed to get all todos: %v", err)
		}
		var checkTitle bool
		if todo1.Title != todoUpdate.Title &&
			todo1.Desc != todoUpdate.Desc &&
			todo1.DoneAt != todoUpdate.DoneAt {
			checkTitle = true
		}

		assert.True(t, checkTitle, todo1.ID)
	})

	//case 5 Delete Todo
	t.Run("DeleteTodoByID", func(t *testing.T) {
		// Tạo một Todo mới để xóa sau
		todo := Todo{
			Title:     "Test Todo to Delete",
			Desc:      "Test Description to Delete",
			Done:      false,
			CreatedAt: time.Now(),
			DoneAt:    nil,
		}

		createdTodo, err := db.CreateTodoDB(ctx, todo)
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		assert.NotEmpty(t, createdTodo.ID, "Todo ID should not be empty")
		assert.Equal(t, createdTodo.Title, todo.Title)
		assert.Equal(t, createdTodo.Desc, todo.Desc)

		err = db.DeleteTodoByIdDB(ctx, createdTodo.ID)
		if err != nil {
			t.Fatalf("Failed to delete todo: %v", err)
		}

		_, err = db.GetTodoByIdDB(ctx, createdTodo.ID)
		if err == nil {
			t.Fatalf("Todo with ID %s was not deleted", createdTodo.ID)
		}
	})

	//case 6 Delete Todo
	t.Run("ChangeStatusById", func(t *testing.T) {
		todo := Todo{
			Title:     "Test Todo for Status Change",
			Desc:      "Test Description for Status Change",
			Done:      false,
			CreatedAt: time.Now(),
			DoneAt:    nil,
		}
		createdTodo, err := db.CreateTodoDB(ctx, todo)
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		assert.NotEmpty(t, createdTodo.ID, "Todo ID should not be empty")
		assert.Equal(t, createdTodo.Done, todo.Done)

		err = db.ChangeStatusDB(ctx, createdTodo.ID)
		if err != nil {
			t.Fatalf("Failed to change status of todo: %v", err)
		}

		changeStatusTodo, err := db.GetTodoByIdDB(ctx, createdTodo.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve updated todo: %v", err)
		}

		assert.Equal(t, changeStatusTodo.Done, true, "Todo status should be updated to true")

		err = db.ChangeStatusDB(ctx, createdTodo.ID)
		if err != nil {
			t.Fatalf("Failed to change status of todo: %v", err)
		}

		changeStatusTodo, err = db.GetTodoByIdDB(ctx, createdTodo.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve updated todo: %v", err)
		}
		assert.Equal(t, changeStatusTodo.Done, false, "Todo status should be updated to false")
	})
}
