package main

type Todo struct {
	ID    int        `json:"id"`
	Name  string     `json:"name"`
	Items []TodoItem `json:"items"`
}

type TodoItem struct {
	ID      int      `json:"id"`
	Content []string `json:"content"`
	Done    bool     `json:"done"`
	TodoID  int      `json:"todo_id"`
}
