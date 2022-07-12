package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type List struct {
	Id          int       `json:"id" db:"id"`
	ListName    string    `json:"listName" db:"list_name"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

var db *sqlx.DB

func init() {
	d, err := sqlx.Open("mysql", "root:root@/todolist?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	db = d
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/lists", allList)
	mux.HandleFunc("/list/", listById)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}

func errJson(w http.ResponseWriter, m string, status int) {
	w.WriteHeader(status)
	err := struct {
		Msg string `json:"errorMsg"`
	}{
		Msg: m,
	}
	json.NewEncoder(w).Encode(err)
}

func allList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		allList := []List{}
		if err := db.Select(&allList, "select * from list"); err != nil {
			errJson(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(allList)
	case http.MethodPost:
		createdList := List{}
		if err := json.NewDecoder(r.Body).Decode(&createdList); err != nil {
			errJson(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := db.NamedExec("insert into list(list_name, title, description) value(:list_name, :title, :description)", createdList)
		if err != nil {
			errJson(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, _ := res.LastInsertId()
		if err := db.Get(&createdList, "select * from list where id=?", id); err != nil {
			errJson(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(createdList)
	default:
		errJson(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func listById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/list/"))
	if err != nil {
		errJson(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		list := List{}
		if err := db.Get(&list, "select * from list where id=?", id); err != nil {
			errJson(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(list)
	case http.MethodPut:
		updatedList := List{}
		if err := json.NewDecoder(r.Body).Decode(&updatedList); err != nil {
			errJson(w, err.Error(), http.StatusBadRequest)
			return
		}
		if updatedList.ListName != "" {
			_, err := db.Exec("update list set list_name=?", updatedList.ListName)
			if err != nil {
				errJson(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if updatedList.Title != "" {
			db.Exec("update list set title=?", updatedList.Title)
			if err != nil {
				errJson(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if updatedList.Description != "" {
			db.Exec("update list set description=?", updatedList.Description)
			if err != nil {
				errJson(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if updatedList.Status != "" {
			db.Exec("update list set status=?", updatedList.Status)
			if err != nil {
				errJson(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if err := db.Get(&updatedList, "select * from list where id=?", id); err != nil {
			errJson(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(updatedList)
	case http.MethodDelete:
		deletedList := List{}
		if err := db.Get(&deletedList, "select * from list where id=?", id); err != nil {
			errJson(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err := db.Exec("delete from list where id=?", id)
		if err != nil {
			errJson(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(deletedList)
	default:
		errJson(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
