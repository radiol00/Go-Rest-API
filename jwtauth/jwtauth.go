package jwtauth

import (
	"database/sql"
	"encoding/json"
	"errors"
	// "fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"time"
)

// JwtController struct should contain database to authenticate users with, use it to populate endpoints
type JwtController struct {
	MYSQLDB *sql.DB
}

type user struct {
	ID       int
	Login    string `json:"login"`
	Password string `json:"password"`
}

type tokens struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

var db *sql.DB

// Authenticate given endpoint using JWT
func Authenticate(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		endpoint(w, r)
	})
}

// PopulateRouter fills given Router with its routes
func (ctrl *JwtController) PopulateRouter(router *mux.Router) {
	db = ctrl.MYSQLDB
	_, err := db.Query("CREATE TABLE IF NOT EXISTS Users(id int NOT NULL AUTO_INCREMENT, login varchar(50), password varchar(50), PRIMARY KEY (id));")
	if err != nil {
		panic(err.Error())
	}

	os.Setenv("JWT_SECRET", "ALAMAKOTA")

	router.HandleFunc("/login", login).Methods("POST")
}

func login(w http.ResponseWriter, r *http.Request) {
	var u user

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := u.getID()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	u.ID = id
	token, err := generateToken(u.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens{Access: token, Refresh: "NOT-IMPLEMENTED"})
}

func (u *user) getID() (int, error) {
	rows, err := db.Query("SELECT * FROM Users WHERE login='" + u.Login + "' AND password='" + u.Password + "';")
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	for rows.Next() {
		var usr user
		err = rows.Scan(&usr.ID, &usr.Login, &usr.Password)
		if err != nil {
			return -1, err
		}
		return usr.ID, nil
	}

	return -1, errors.New("Not Found")
}

func generateToken(userid int) (string, error) {
	var err error

	claims := jwt.MapClaims{}
	claims["id"] = userid
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return token, nil
}
