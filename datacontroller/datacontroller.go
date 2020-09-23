package datacontroller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golangrestapi/jwtauth"
	"net/http"
	"strings"
)

// DataController : contains router that handles data oriented endpoints
type DataController struct {
	MYSQLDB *sql.DB
}

var db *sql.DB

type myData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// PopulateRouter fills given Router with its routes
func (ctrl *DataController) PopulateRouter(router *mux.Router) {
	db = ctrl.MYSQLDB
	_, err := db.Query("CREATE TABLE IF NOT EXISTS MyData(id int NOT NULL AUTO_INCREMENT, name varchar(50), PRIMARY KEY (id));")
	if err != nil {
		panic(err.Error())
	}

	router.Handle("/", jwtauth.Authenticate(getAllData)).Methods("GET")
	router.Handle("/", jwtauth.Authenticate(createData)).Methods("POST")
	router.Handle("/{id}", jwtauth.Authenticate(getData)).Methods("GET")
	router.Handle("/{id}", jwtauth.Authenticate(deleteData)).Methods("DELETE")
	router.Handle("/{id}", jwtauth.Authenticate(updateData)).Methods("PUT")
}

func checkIfRowExists(id string) (myData, bool, error) {
	rows, err := db.Query("SELECT * FROM MyData WHERE id='" + id + "';")
	if err != nil {
		return myData{}, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var value myData
		err = rows.Scan(&value.ID, &value.Name)
		return value, true, err
	}

	return myData{}, false, nil
}

func getData(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	value, exists, err := checkIfRowExists(id)

	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(value)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func getAllData(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM MyData")
	defer rows.Close()

	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	d := []myData{}
	for rows.Next() {
		var value myData

		err = rows.Scan(&value.ID, &value.Name)

		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		d = append(d, value)
	}

	json.NewEncoder(w).Encode(d)
}

func createData(w http.ResponseWriter, r *http.Request) {
	var value myData
	json.NewDecoder(r.Body).Decode(&value)
	value.Name = strings.TrimSpace(value.Name)

	if value.Name == "" {
		fmt.Println("myData : Bad name value")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := db.Query("INSERT INTO MyData(name) VALUES ('" + value.Name + "')")
	defer result.Close()

	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func deleteData(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	// rows, err := db.Query("SELECT * FROM MyData WHERE id='" + id + "'")
	_, exists, err := checkIfRowExists(id)

	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	rows, err := db.Query("DELETE FROM MyData WHERE id='" + id + "';")
	rows.Close()

	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func updateData(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	_, exists, err := checkIfRowExists(id)

	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var value myData
	json.NewDecoder(r.Body).Decode(&value)
	if value.Name == "" {
		fmt.Println("myData : Bad name value")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rows, err := db.Query("UPDATE MyData SET name='" + value.Name + "' WHERE id='" + id + "';")
	rows.Close()

	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func login(w http.ResponseWriter, r *http.Request) {

}
