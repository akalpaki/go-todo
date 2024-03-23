package main

import "database/sql"

const MIGRATION_TEMP = `
create table if not exists todo (id integer not null primary key autoincrement, name text not null);
create table if not exists todo_item (itemNo integer not null, content text not null, done boolean not null, todo_id integer not null, foreign key (todo_id) references todo (id) on delete cascade);
`

func runMigration(db *sql.DB) {
	_, err := db.Exec(MIGRATION_TEMP)
	if err != nil {
		panic("Unable to migrate database!")
	}
}

func main() {
	cfg := LoadConfig()

	runMigration(cfg.DB)

	storer := NewStorer(cfg.DB)

	app := NewRestServer(cfg, storer)
	app.Run()
}
