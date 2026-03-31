package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golangcollege/sessions"
	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB
var testApp *application

func TestMain(m *testing.M) {
	fmt.Println("Starting tests...")
	var err error
	testDB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	if err := testDB.Ping(); err != nil {
		panic(err)
	}

	testApp = setupApp(testDB)
	if err = setUpTestShema(testDB); err != nil {
		panic(err)
	}

	code := m.Run()
	testDB.Close()
	fmt.Println("Finished tests...")
	os.Exit(code)
}

func setupApp(db *sql.DB) *application {
	session := sessions.New([]byte("test-secret-key"))
	session.Lifetime = 24 * time.Hour

	return &application{
		errorLog:    log.New(io.Discard, "", 0),
		infoLog:     log.New(io.Discard, "", 0),
		userRepo:    NewSQLUserRepository(db),
		postRepo:    NewSQLPostRepository(db),
		templateDir: "./templates",
		publicPath:  "./public",
		session:     session,
	}

}

func setUpTestShema(db *sql.DB) error {
	schema := `
	CREATE TABLE users (
   id INTEGER PRIMARY KEY AUTOINCREMENT,
   name TEXT NOT NULL,
   email TEXT NOT NULL UNIQUE,
   hashed_password TEXT NOT NULL,
   created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE profile (
     user_id INTEGER PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
     avatar TEXT NOT NULL,
     created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE posts (
   id INTEGER PRIMARY KEY AUTOINCREMENT,
   url TEXT NOT NULL,
   title TEXT NOT NULL UNIQUE,
   user_id INTEGER REFERENCES users(user_id) ON DELETE CASCADE,
   created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE comments (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  body TEXT NOT NULL,
  user_id INTEGER REFERENCES users(user_id) ON DELETE CASCADE,
  post_id INTEGER REFERENCES posts(post_id) ON DELETE CASCADE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE votes (
   user_id INTEGER REFERENCES users(user_id) ON DELETE CASCADE,
   post_id INTEGER REFERENCES posts(post_id) ON DELETE CASCADE,
   created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
   PRIMARY KEY (user_id, post_id)
);
	`
	_, err := db.Exec(schema)
	return err
}

func cleanupTestData(t *testing.T) {
	tables := []string{
		"profile",
		"users",
		"posts",
		"comments",
		"votes",
	}

	for _, table := range tables {
		_, err := testDB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		assert.NoError(t, err)
	}
}
