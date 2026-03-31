package main

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLUserRepository_CreateUser(t *testing.T) {
	defer cleanupTestData(t)

	repo := NewSQLUserRepository(testDB)

	userID, err := repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)
	assert.Nil(t, err)
	assert.Greater(t, userID, 0)

	user, err := repo.GetUserByEmail("john@doe.com")
	assert.Nil(t, err)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)

	userID, err = repo.CreateUser(
		"John Doe",
		"john2@doe.com",
		generateString(73),
		"avatar",
	)
	assert.Error(t, err)
	assert.Equal(t, userID, 0)

}

func TestSQLUserRepository_CreateUser_DuplicateEmail(t *testing.T) {
	defer cleanupTestData(t)

	repo := NewSQLUserRepository(testDB)

	userID, err := repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)
	assert.Nil(t, err)
	assert.Greater(t, userID, 0)

	_, err = repo.CreateUser(
		"John Doe",
		"john@doe.com",
		generateString(73),
		"avatar",
	)
	assert.Error(t, err)
}

func TestSQLUserRepository_Authenticate(t *testing.T) {
	defer cleanupTestData(t)

	repo := NewSQLUserRepository(testDB)

	userID, err := repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)

	assert.Nil(t, err)
	assert.Greater(t, userID, 0)

	authUserID, err := repo.Authenticate("john@doe.com", "testpassword")
	assert.NoError(t, err)
	assert.Equal(t, userID, authUserID)
}

func TestSQLUserRepository_Authenticate_WrongPassword(t *testing.T) {
	defer cleanupTestData(t)

	repo := NewSQLUserRepository(testDB)

	userID, err := repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)

	assert.Nil(t, err)
	assert.Greater(t, userID, 0)

	_, err = repo.Authenticate("john@doe.com", "wrongpassword")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredential, err)

}

func generateString(n int) string {
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i] = 'a'
	}
	return string(buf)
}

func TestNewSQLUserRepository(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		db   *sql.DB
		want UserRepository
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSQLUserRepository(tt.db)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("NewSQLUserRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}
