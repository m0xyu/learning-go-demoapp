package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB

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
	if err = setUpTestShema(testDB); err != nil {
		panic(err)
	}

	code := m.Run()
	testDB.Close()
	fmt.Println("Finished tests...")
	os.Exit(code)
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
	_, err := testDB.Exec("DELETE FROM profile")
	assert.NoError(t, err)
	_, err = testDB.Exec("DELETE FROM users")
	assert.NoError(t, err)
	_, err = testDB.Exec("DELETE FROM posts")
	assert.NoError(t, err)
	_, err = testDB.Exec("DELETE FROM comments")
	assert.NoError(t, err)
	_, err = testDB.Exec("DELETE FROM votes")
	assert.NoError(t, err)
}
