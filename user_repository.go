package main

import (
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredential = errors.New("invalid credential")

type UserRepository interface {
	CreateUser(name, email, plainPassword, avatar string) (int, error)
	GetUserByEmail(email string) (*User, error)
	GetUsers() ([]User, error)
	Authenticate(email, password string) (int, error)
}

type SQLUserRepository struct {
	db *sql.DB
}

// NewSQLUserRepository create new UserRepository type
func NewSQLUserRepository(db *sql.DB) UserRepository {
	return &SQLUserRepository{
		db: db,
	}
}

func (r *SQLUserRepository) CreateUser(name, email, plainPassword, avatar string) (int, error) {
	ctx := context.Background()

	// transition開始
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	// return（エラー終了）したとき、自動的にロールバックが発動
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO users (name, email, hashed_password) VALUES (?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// パスワードのハッシュ化
	hp, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// sqlの結果を受け取る
	result, err := stmt.Exec(name, email, string(hp))
	if err != nil {
		return 0, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	profileStm, err := tx.PrepareContext(ctx, `INSERT INTO profile (user_id, avatar) VALUES( ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer profileStm.Close()

	_, err = profileStm.Exec(userID, avatar)
	if err != nil {
		return 0, err
	}

	// 確定
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return int(userID), nil
}

func (r *SQLUserRepository) GetUserByEmail(email string) (*User, error) {
	stmt := `SELECT u.id, u.name, u.email,  u.hashed_password, u.created_at, p.avatar FROM users u 
	INNER JOIN profile p ON u.id = p.user_id WHERE u.email = ?`

	row := r.db.QueryRow(stmt, email)

	var user User
	// 宣言したuserに値を直接書き込む
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.HashedPassword, &user.CreatedAt, &user.Profile.Avatar)
	if err != nil {
		return nil, err
	}
	user.Profile.UserID = user.ID
	return &user, nil
}

func (r *SQLUserRepository) Authenticate(email, password string) (int, error) {
	user, err := r.GetUserByEmail(email)
	if err != nil {
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredential
		}
		return 0, err
	}
	return user.ID, nil
}

func (r *SQLUserRepository) GetUsers() ([]User, error) {
	stmt := `SELECT id, name, email,  hashed_password, created_at FROM users`
	rows, err := r.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	// 次の行があればtrue、なければfalse
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.HashedPassword,
			&user.CreatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	// DBの接続などエラー判定
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
