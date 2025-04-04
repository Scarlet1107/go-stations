package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/TechBowl-japan/go-stations/model"
)

// A TODOService implements CRUD of TODO entities.
type TODOService struct {
	db *sql.DB
}

// NewTODOService returns new TODOService.
func NewTODOService(db *sql.DB) *TODOService {
	return &TODOService{
		db: db,
	}
}

// CreateTODO creates a TODO on DB.
// CreateTODOはTODOServeiceのsという構造体と紐づいている
// ctx, subject, description	普通の引数（呼び出し時に渡す値）
// (*model.TODO, error)	関数の返り値。保存したTODOとエラー情報
func (s *TODOService) CreateTODO(ctx context.Context, subject, description string) (*model.TODO, error) {
	const (
		insert  = `INSERT INTO todos(subject, description) VALUES(?, ?)`
		confirm = `SELECT subject, description, created_at, updated_at FROM todos WHERE id = ?`
	)

	// ExecContextはDBに対してSQLを実行するための関数。（Select以外）
	res, err := s.db.ExecContext(ctx, insert, subject, description)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Selectで挿入したTODOを取得
	row := s.db.QueryRowContext(ctx, confirm, id)

	var todo model.TODO
	todo.ID = id

	err = row.Scan(
		&todo.Subject,
		&todo.Description,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &todo, nil
}

// ReadTODO reads TODOs on DB.
func (s *TODOService) ReadTODO(ctx context.Context, prevID, size int64) ([]*model.TODO, error) {
	const (
		read       = `SELECT id, subject, description, created_at, updated_at FROM todos ORDER BY id DESC LIMIT ?`
		readWithID = `SELECT id, subject, description, created_at, updated_at FROM todos WHERE id < ? ORDER BY id DESC LIMIT ?`
	)

	var rows *sql.Rows
	var err error

	if prevID == 0 {
		rows, err = s.db.QueryContext(ctx, read, size)
	} else {
		rows, err = s.db.QueryContext(ctx, readWithID, prevID, size)
	}
	if err != nil {
		return nil, err
	}
	// deferは関数が終了する時に実行される
	//
	defer rows.Close()

	todos := []*model.TODO{}

	for rows.Next() {
		var todo model.TODO
		err := rows.Scan(
			&todo.ID,
			&todo.Subject,
			&todo.Description,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, &todo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

// UpdateTODO updates the TODO on DB.
func (s *TODOService) UpdateTODO(ctx context.Context, id int64, subject, description string) (*model.TODO, error) {
	const (
		update  = `UPDATE todos SET subject = ?, description = ? WHERE id = ?`
		confirm = `SELECT subject, description, created_at, updated_at FROM todos WHERE id = ?`
	)

	res, err := s.db.ExecContext(ctx, update, subject, description, id)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, &model.ErrNotFound{}
	}

	row := s.db.QueryRowContext(ctx, confirm, id)
	var todo model.TODO
	todo.ID = id

	err = row.Scan(
		&todo.Subject,
		&todo.Description,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &todo, nil
}

// DeleteTODO deletes TODOs on DB by ids.
func (s *TODOService) DeleteTODO(ctx context.Context, ids []int64) error {
	const deleteFmt = `DELETE FROM todos WHERE id IN (?%s)`

	if len(ids) == 0 {
		return nil
	}

	// ② クエリ文字列の動的構築：?を増やす
	placeholders := strings.Repeat(",?", len(ids)-1) // 例: ",?,?"
	query := fmt.Sprintf(deleteFmt, placeholders)    // 完成: "DELETE FROM todos WHERE id IN (?, ?, ...)"

	// ③ []int64 → []interface{} に変換（ExecContextに渡すため）
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	// ④ SQL実行
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	// ⑤ 削除件数の確認
	rowAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowAffected == 0 {
		return &model.ErrNotFound{}
	}

	return nil
}
