package main

type Todo struct {
	ID    int
	Name  string
	Items []TodoItem
}

type TodoItem struct {
	ID      int
	Content []string
	Done    bool
	TodoID  int
}
