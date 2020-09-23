package jwtauth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"strconv"
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
	rows, err := db.Query("CREATE TABLE IF NOT EXISTS Users(id int NOT NULL AUTO_INCREMENT, login varchar(50), password varchar(50), PRIMARY KEY (id));")
	if err != nil {
		panic(err.Error())
	}
	rows.Close()

	rows, err = db.Query("CREATE TABLE IF NOT EXISTS Tokens(id int NOT NULL AUTO_INCREMENT, user_id int NOT NULL, token varchar(255), exp DATETIME NOT NULL, PRIMARY KEY (id));")
	if err != nil {
		panic(err.Error())
	}
	rows.Close()

	rows, err = db.Query("CREATE EVENT IF NOT EXISTS `golangrestapidb`.`jwtExpiringEvent` ON SCHEDULE EVERY 1 MINUTE COMMENT 'I delete expired JWT Tokens' DO BEGIN DELETE FROM Tokens WHERE exp < NOW(); END")
	if err != nil {
		panic(err.Error())
	}
	rows.Close()

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
	claims := jwt.MapClaims{}
	claims["id"] = userid
	in15mins := time.Now().Add(time.Minute * 2)
	claims["exp"] = in15mins.Unix()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	// add token to db
	rows, err := db.Query("INSERT INTO Tokens(user_id, token, exp) VALUES ('" + strconv.Itoa(userid) + "', '" + token + "', '" + in15mins.Format("2006-01-02 15:04:05") + "')")
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	rows.Close()

	return token, nil
}
