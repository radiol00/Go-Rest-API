package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"golangrestapi/datacontroller"
	"golangrestapi/jwtauth"
	"net/http"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/golangrestapidb")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	router := mux.NewRouter()
	datacontroller := datacontroller.DataController{MYSQLDB: db}
	jwtcontroller := jwtauth.JwtController{MYSQLDB: db}
	datacontroller.PopulateRouter(router)
	jwtcontroller.PopulateRouter(router)

	server := &http.Server{Addr: ":8080", Handler: router}

	fmt.Println("SERVING")
	go server.ListenAndServe()

	var command string
	for command != "kill" && command != "k" {
		fmt.Println("type 'kill' or 'k' to stop serving")
		fmt.Scanln(&command)
	}
	server.Close()
	fmt.Println("ENDED")
}
