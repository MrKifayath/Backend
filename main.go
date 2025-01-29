package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
)

var db *pgx.Conn

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "About Section\n")
	fmt.Fprintf(w, "This is a portfolio website built with Go!")
}

func connectToDB() {
	var err error
	db, err = pgx.Connect(context.Background(), "postgres://postgres:Kifayath@localhost:5432/demo")
	if err != nil {
		log.Fatalf("Unable to connect to the database: %v\n", err)
	}
	fmt.Println("Connected to database")
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		name := r.FormValue("name")
		message := r.FormValue("message")
		fmt.Fprintf(w, "Received message from %s: %s", name, message)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Only POST method is allowed.")
	}
}

func getProjectsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(context.Background(), "SELECT id, name, description, link FROM projects")
	if err != nil {
		http.Error(w, "Something unexpected happened", http.StatusInternalServerError)
	}
	defer rows.Close()

	var projects []map[string]interface{}
	for rows.Next() {
		var id int
		var name, description, link string
		err := rows.Scan(&id, &name, &description, &link)
		if err != nil {
			http.Error(w, "Couldn't scan the rows", http.StatusInternalServerError)
		}
		projects = append(projects, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"link":        link,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)

}

func createProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
	}

	var project map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&project)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
	}
	name, _ := project["name"].(string)
	description, _ := project["description"].(string)
	link, _ := project["link"].(string)

	_, err = db.Exec(context.Background(), "INSERT INTO projects (id, name, description, link) VALUES ($1, $2, $3)", name, description, link)
	if err != nil {
		http.Error(w, "Failed to insert project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Project created successfully")
}

func updateProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
	}

	var project map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&project)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
	}

	id, _ := project["id"].(float64)
	name, _ := project["name"].(string)
	description, _ := project["description"].(string)
	link, _ := project["link"].(string)

	_, err = db.Exec(context.Background(), "UPDATE projects SET name=$1, description=$2, link=$3, id==$4", int(id), name, description, link)
	if err != nil {
		http.Error(w, "Failed to update project", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Project updated successfully")
}

func main() {
	connectToDB()
	defer db.Close(context.Background())

	fmt.Println("Server is running...")
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/projects", getProjectsHandler)
	http.HandleFunc("/projects/create", createProjectHandler)
	http.HandleFunc("/projects/update", updateProjectHandler)

	fmt.Println("Server is running on localhost:3002")
	http.ListenAndServe(":3002", nil)
}
