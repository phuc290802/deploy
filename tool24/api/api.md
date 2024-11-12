# API
- Write HTTP API to manage todos
- Write mock `TodoService`
- Write unit test for HTTP API

## Tech requirements
- Use `net/http` ServerMux for routing
- Use `testing` for unit test
- Folder structure:
  + `main.go`: bootstrap application
  + `todo.go`: declare `Todo`, `TodoService`
  + `api.go`: implement todo apis
  + `api_test.go`: unit tests for todo apis
 
## HTTP API
```
# get all todos
GET /todos
# get one todo
GET /todo/{id}
# create a todo
POST /todo
# update a todo
PATCH /todo/{id}
# delete a todo
DELETE /todo/{id}
```

## TodoService
```go
struct TodoService {
  todos Todo[]
}

# method to get all
# method to get one
# method to create
# method to update
# method to delete
```
## Unit test
- Test http API using mock `TodoService`

## References
- https://pkg.go.dev/net/http#Handle
- https://gobyexample.com/http-server
- https://gobyexample.com/testing-and-benchmarking



