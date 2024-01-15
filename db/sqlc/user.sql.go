// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: user.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const countUsers = `-- name: CountUsers :one
SELECT count(*) FROM users
`

func (q *Queries) CountUsers(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, countUsers)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (
  username,
  password,
  full_name,
  email
) VALUES (
  $1, $2, $3, $4
) RETURNING id, username, full_name, email, password, email_verified_at, password_changed_at, created_at, updated_at
`

type CreateUserParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.Username,
		arg.Password,
		arg.FullName,
		arg.Email,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.FullName,
		&i.Email,
		&i.Password,
		&i.EmailVerifiedAt,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1
`

func (q *Queries) DeleteUser(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteUser, id)
	return err
}

const getUser = `-- name: GetUser :one
SELECT id, username, full_name, email, password, email_verified_at, password_changed_at, created_at, updated_at FROM users
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetUser(ctx context.Context, id int64) (User, error) {
	row := q.db.QueryRow(ctx, getUser, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.FullName,
		&i.Email,
		&i.Password,
		&i.EmailVerifiedAt,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, username, full_name, email, password, email_verified_at, password_changed_at, created_at, updated_at FROM users
WHERE email = $1 LIMIT 1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.FullName,
		&i.Email,
		&i.Password,
		&i.EmailVerifiedAt,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
SELECT id, username, full_name, email, password, email_verified_at, password_changed_at, created_at, updated_at FROM users
WHERE username = $1 LIMIT 1
`

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByUsername, username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.FullName,
		&i.Email,
		&i.Password,
		&i.EmailVerifiedAt,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listUsers = `-- name: ListUsers :many
SELECT id, username, full_name, email, password, email_verified_at, password_changed_at, created_at, updated_at FROM users
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListUsersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error) {
	rows, err := q.db.Query(ctx, listUsers, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.FullName,
			&i.Email,
			&i.Password,
			&i.EmailVerifiedAt,
			&i.PasswordChangedAt,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateUser = `-- name: UpdateUser :one
UPDATE users
SET
  password = COALESCE($1, password),
  password_changed_at = COALESCE($2, password_changed_at),
  full_name = COALESCE($3, full_name),
  email = COALESCE($4, email),
  updated_at = now()
WHERE
  id = $5
RETURNING id, username, full_name, email, password, email_verified_at, password_changed_at, created_at, updated_at
`

type UpdateUserParams struct {
	Password          pgtype.Text        `json:"password"`
	PasswordChangedAt pgtype.Timestamptz `json:"password_changed_at"`
	FullName          pgtype.Text        `json:"full_name"`
	Email             pgtype.Text        `json:"email"`
	ID                int64              `json:"id"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUser,
		arg.Password,
		arg.PasswordChangedAt,
		arg.FullName,
		arg.Email,
		arg.ID,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.FullName,
		&i.Email,
		&i.Password,
		&i.EmailVerifiedAt,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
