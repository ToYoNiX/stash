package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ToYoNiX/stash/internal/app"
	"github.com/ToYoNiX/stash/internal/containers"
	"github.com/ToYoNiX/stash/internal/items"
)

func main() {
	// SQLite ignores FOREIGN KEY constraints unless they are turned on
	// for every connection, which is what _foreign_keys=on does.
	db, err := sql.Open("sqlite3", "./db.sqlite3?_foreign_keys=on")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	entries, err := os.ReadDir("migrations")
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join("migrations", entry.Name()))
		if err != nil {
			log.Fatal(err)
		}

		sqlStmt := string(content)
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Fatal(err)
		}
	}

	cName := "first"
	cDesc := "empty"
	c := containers.Container{Name: &cName, Description: &cDesc}
	cId, err := containers.Insert(db, &c)
	if err != nil {
		log.Fatal(err)
	}

	iName := "POCO"
	iDesc := "My old phone"
	iAmount := 1
	i := items.Item{Name: &iName, Description: &iDesc, Amount: &iAmount, ContainerId: &cId}
	_, err = app.AddItem(db, &i)
	if err != nil {
		log.Fatal(err)
	}
}
