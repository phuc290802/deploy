# DB Demo
- Create `todo` schema in PostgreSQL
- Write Go lang code for `TodoService` with `todo` schema
- Use migration tool to manage schema version

## Schema
```
id: string
title: string
desc: string
done: bool
created_at: timestamp
```

Schema migration
- Add `done_at: timestamp` column

## Tech requirements
- Use `pgx` for connection to PostgreSQL
- Use `migrate` for schema migration 
- Folder structure:
+ `main.go`: bootstrap application
+ `db.go`: declare Db struct to wrap and connect to PostgreSQL DB
+ `todo.go`: declare `Todo` and implement `TodoService`
+ `todo_test.go`: unit tests for `TodoService`

## References
- https://github.com/jackc/pgx
- https://github.com/golang-migrate/migrate
