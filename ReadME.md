# Collaborative Story Assignment

## Instructions To Run The Application: 

### 1. Using Docker 

1. Run `docker-compose build` 
2. Run `docker-compose up` (in root directory where **docker-compose.yml** is located).
3. API Requests can be made at `http://localhost:8080/`


### 2. Locally (Without Docker)

1. Make sure **MySQL** server is installed and running on local machine
2. Set ENV variable `DB_URL`
    ```
    export DB_URL="mysql://[user]:[password]@tcp([db_host]:3306)/[db_name]"
    ```
3. Optionally set ENV variable `LOGS_ENABLE=1` to enable logging. 
4. Build using `go build -o bin/server main.go`
5. Run `./bin/server` 
6. OR  Run `go run main.go` directly.
7. API Requests can be made at `http://localhost:8080/`


## Things I would have done if I had more time: 
1. Better test coverage with more Unit and Functional tests written.
2. Better metrics calculation implementation using some library.
3. Detailed profiling of the application.
4. Better documentation with docstrings. 
