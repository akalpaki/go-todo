package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const BANNER = `
████████╗ ██████╗ ██████╗  ██████╗ 
╚══██╔══╝██╔═══██╗██╔══██╗██╔═══██╗
   ██║   ██║   ██║██║  ██║██║   ██║
   ██║   ██║   ██║██║  ██║██║   ██║
   ██║   ╚██████╔╝██████╔╝╚██████╔╝
   ╚═╝    ╚═════╝ ╚═════╝  ╚═════╝
-----------------------------------
A minimalistic CLI todo-app (may get extended to include web UI)
Purpose:
- todo get: retrieve stored todo items
- todo add: add an item to the todo list
- todo delete: delete stored todo items
`

const MIGRATION_TEMP = `
create table if not exists todo (id integer not null primary key autoincrement, todo text);
`

func addTodo(db *sql.DB, todo string) error {
	_, err := db.Exec("insert into todo (todo) values (?)", todo)
	return err
}

func getTodo(db *sql.DB) (string, error) {
	res := ""
	rows, err := db.Query("select todo from todo")
	if err != nil {
		return "", err
	}
	for rows.Next() {
		newStr := ""
		if err := rows.Scan(&newStr); err != nil {
			return "", err
		}
		res += "\n"
		res += newStr
	}
	return res, nil
}

func deleteTodo(db *sql.DB) error {
	_, err := db.Exec("drop table if exists todo")
	return err
}

func main() {
	db, err := sql.Open("sqlite3", "./todo.db")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer db.Close()

	_, err = db.Exec(MIGRATION_TEMP)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	add := flag.NewFlagSet("add", flag.ExitOnError)
	get := flag.NewFlagSet("get", flag.ExitOnError)
	delete := flag.NewFlagSet("delete", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Print(BANNER)
		return
	}

	switch os.Args[1] {
	case "add":
		add.Parse(os.Args[2:])
		if err := addTodo(db, strings.Join(os.Args[2:], " ")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Todo Added")
	case "get":
		get.Parse(os.Args[2:])
		todos, err := getTodo(db)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(todos)
	case "delete":
		delete.Parse(os.Args[2:])
		if err := deleteTodo(db); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Todo Deleted")
	}
}
