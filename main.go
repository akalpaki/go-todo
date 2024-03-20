package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

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
- todo create: create an item to the todo list
- todo delete: delete stored todo items
`

const MIGRATION_TEMP = `
create table if not exists todo (id integer not null primary key autoincrement, name text not null);
create table if not exists todo_item (id integer not null primary key autoincrement, content text, done boolean not null, todo_id integer not null, foreign key (todo_id) references todo (id) on delete cascade);
`

type todo struct {
	id    int
	name  string
	items []todoItem
}

type todoItem struct {
	id      int
	content []string
	done    bool
	todoID  int
}

// createTodo creates a new Todo list. If todo items are passed, they are added to the list.
func createTodo(ctx context.Context, db *sql.DB, td todo) (todo, error) {
	addItem, prepErr := db.Prepare("insert into todo_item(content, done, todo_id) values (?, ?, ?)") // TODO: look into preparing these once somewhere, maybe sync once?
	if prepErr != nil {
		return todo{}, prepErr
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return todo{}, err
	}
	txAddItem := tx.Stmt(addItem)

	_, err = tx.Exec("insert into todo (name) values ?", td.name)
	if err != nil {
		tx.Rollback()
		return todo{}, err
	}

	if len(td.items) > 0 {
		for i, v := range td.items {
			res, err := txAddItem.Exec(v.content, v.done, v.todoID)
			if err != nil {
				tx.Rollback()
				return todo{}, err
			}
			id, err := res.LastInsertId()
			if err != nil {
				tx.Rollback()
				return todo{}, err
			}
			td.items[i].id = int(id)
		}
	}

	err = tx.Commit()
	if err != nil {
		return todo{}, err
	}

	return td, nil
}

// addTodoItems adds any number of todo items for a specific todo list
func addTodoItems(ctx context.Context, db *sql.DB, items []todoItem) ([]todoItem, error) {
	addItem, err := db.Prepare("insert into todo_item (content, done, todo_id) values (?, ?, ?)")
	if err != nil {
		return nil, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	txAddItem := tx.Stmt(addItem)
	for i, v := range items {
		res, err := txAddItem.ExecContext(ctx, v.content, v.done, v.todoID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		items[i].id = int(id)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return items, err
}

// updateTodoItem can be used to update an individual item from a todo list
func updateTodoItem(ctx context.Context, db *sql.DB, item todoItem) (todoItem, error) {
	_, err := db.ExecContext(ctx, "update todo_item (content, done) where id = ?", item.id)
	if err != nil {
		return todoItem{}, err
	}
	return item, err
}

func deleteTodo(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, "delete from todo where id = ?", id)
	return err
}

func deleteTodoItem(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, "delete from todo_item where id = ?", id)
	return err
}

func getTodo(ctx context.Context, db *sql.DB, id int) (todo, error) {
	var res todo

	tdRow := db.QueryRowContext(ctx, "select * from todo where id = ?", id)
	if err := tdRow.Scan(&res); err != nil {
		return todo{}, err
	}

	itemRows, err := db.QueryContext(ctx, "select * from todo_item where todo_id = ?", id)
	if err != nil {
		return todo{}, err
	}

	for itemRows.Next() {
		var item todoItem
		if err := itemRows.Scan(&item); err != nil {
			return todo{}, err
		}
		res.items = append(res.items, item)
	}

	return res, nil
}

func main() {
	ctx := context.Background()
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

	create := flag.NewFlagSet("create", flag.ExitOnError)
	get := flag.NewFlagSet("get", flag.ExitOnError)
	delete := flag.NewFlagSet("delete", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Print(BANNER)
		return
	}

	switch os.Args[1] {
	case "create":
		// create.Parse(os.Args[2:])

		// td := todo{
		// 	name: os.Args[2],
		// }

		// if err := createTodo(ctx, db, strings.Join(os.Args[2:], " ")); err != nil {
		// 	fmt.Println(err)
		// 	os.Exit(1)
		// }
		// fmt.Println("Todo Added")
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
