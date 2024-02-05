package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"todoapp/db"

	_ "github.com/go-sql-driver/mysql"
)

type Todo struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdateTodo struct {
	ID      int    `json:"id"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

var (
	ServiceName = "todo-app"
	port        = "8080"
)

var connectdb = db.ConnectDB()

func getListTodo(w http.ResponseWriter, req *http.Request) {
	var todoList []Todo
	rows, err := connectdb.Query("SELECT * FROM todos")

	if err != nil {
		w.Write([]byte(fmt.Sprintf("Todo list is empty")))
		return
	}

	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Content); err != nil {
			w.Write([]byte(fmt.Sprintf("Todo list is empty 1")))
			return
		}

		todoList = append(todoList, todo)
	}

	result, err := json.Marshal(todoList)

	if err != nil {
		w.Write([]byte(fmt.Sprintf("Todo list is empty 2")))
		return
	}

	w.Write(result)
}

func updateTodo(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)

	if err != nil {
		// Handle the error if JSON marshaling fails
		http.Error(w, "Failed to marshal albums to JSON", http.StatusInternalServerError)
		return
	}

	var todo UpdateTodo
	if errs := json.Unmarshal(data, &todo); errs != nil {
		http.Error(w, "Failed to marshal todo to JSON 1", http.StatusInternalServerError)
		return
	}

	var (
		titleUpdate   string
		contentUpdate string
	)

	connectdb.QueryRow("SELECT title, content FROM todos where id = ?", todo.ID).Scan(&titleUpdate, &contentUpdate)

	if todo.Title != "" {
		titleUpdate = todo.Title
	}

	if todo.Content != "" {
		contentUpdate = todo.Content
	}

	_, err = connectdb.Exec(`
		UPDATE todos
		SET title =  ?, content = ?
		WHERE id = ?`, titleUpdate, contentUpdate, todo.ID)

	if err != nil {
		message := fmt.Sprintf("Failed to marshal albums to JSON 2. %v", err)
		w.Write([]byte(message))
		return
	}

	// Truy vấn bản ghi đã được cập nhật
	updatedTodo := Todo{}
	err = connectdb.QueryRow(`
		SELECT * FROM todos
		WHERE id = ?`, todo.ID).Scan(&updatedTodo.ID, &updatedTodo.Title, &updatedTodo.Content)

	if err != nil {
		message := fmt.Sprintf("Failed to query updated todo: %v", err)
		w.Write([]byte(message))
		return
	}

	result, err := json.Marshal(updatedTodo)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Todo list is empty 4")))
		return
	}

	w.Write(result)
}

func storeTodo(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)

	if err != nil {
		// Handle the error if JSON marshaling fails
		http.Error(w, "Failed to marshal albums to JSON 0", http.StatusInternalServerError)
		return
	}

	var todo Todo
	if errs := json.Unmarshal(data, &todo); errs != nil {
		http.Error(w, "Failed to marshal todo to JSON 1", http.StatusInternalServerError)
		return
	}

	rows, err := connectdb.Exec("INSERT INTO todos (id, title, content) VALUES (null, ?, ?)", todo.Title, todo.Content)

	if err != nil {
		message := fmt.Sprintf("Failed to marshal albums to JSON 2. %v", err)
		http.Error(w, message, http.StatusInternalServerError)
		return
	}

	id, err := rows.LastInsertId()
	var todoList Todo

	row := connectdb.QueryRow("SELECT * FROM todos WHERE id = ?", id)
	if err := row.Scan(&todoList.ID, &todoList.Title, &todoList.Content); err != nil {
		if err == sql.ErrNoRows {
			w.Write([]byte(fmt.Sprintf("Todo list is empty")))
		}
		w.Write([]byte(fmt.Sprintf("Todo list is empty 2")))
	}

	result, err := json.Marshal(todoList)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Todo list is empty 3")))
	}

	w.Write(result)
}

func main() {
	// Create the database if it doesn't exist
	_, err := connectdb.Exec("CREATE DATABASE IF NOT EXISTS todoapp")
	if err != nil {
		log.Fatal(err)
	}

	// Create a sample table
	_, err = connectdb.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(50),
			content VARCHAR(50)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer connectdb.Close()

	http.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(fmt.Sprintf("Hello, Word! adse ok %s", ServiceName)))
	})

	http.HandleFunc("/get-list", getListTodo)
	http.HandleFunc("/update", updateTodo)
	http.HandleFunc("/store", storeTodo)
	http.HandleFunc("/get-todo", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		if id == "" {
			http.Error(w, "id is empty", http.StatusBadRequest)
			return
		}

		var todo Todo
		err = connectdb.QueryRow("SELECT * FROM todos where id = ?", id).Scan(&todo.ID, &todo.Title, &todo.Content)
		if err != nil {
			http.Error(w, "Can't find todo", http.StatusBadRequest)
			return
		}

		result, err := json.Marshal(todo)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		w.Write(result)
	})

	fmt.Println("start service with port: ", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
