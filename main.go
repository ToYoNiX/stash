package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ToYoNiX/stash/internal/web"
)

func main() {
	addr := flag.String("addr", ":8090", "address to listen on")
	dbPath := flag.String("db", "./data/db.sqlite3", "path to the sqlite database file")
	migrations := flag.String("migrations", "./migrations", "path for sql migrations")
	uploadsPath := flag.String("uploads", "./data/uploads", "directory to store uploaded photos")
	flag.Parse()

	err := os.MkdirAll(filepath.Dir(*dbPath), 0o755)
	if err != nil {
		log.Fatal(err)
	}

	// SQLite ignores FOREIGN KEY constraints unless they are turned on
	// for every connection, which is what _foreign_keys=on does.
	db, err := sql.Open("sqlite3", *dbPath + "?_foreign_keys=on")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	entries, err := os.ReadDir(*migrations)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join(*migrations, entry.Name()))
		if err != nil {
			log.Fatal(err)
		}

		sqlStmt := string(content)
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Fatal(err)
		}
	}

	server, err := web.New(db, *uploadsPath)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening on", *addr)
	log.Fatal(http.ListenAndServe(*addr, server.Handler()))
}
