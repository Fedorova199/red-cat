package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Database struct {
	db *sql.DB
}

var ErrDeleted = errors.New("deleted")

func CreateDatabase(db *sql.DB) (*Database, error) {
	databaseStorage := &Database{
		db: db,
	}

	err := databaseStorage.init()
	if err != nil {
		return nil, err
	}

	return databaseStorage, nil
}

func (s *Database) init() error {
	_, err := s.db.Exec("CREATE TABLE IF NOT EXISTS url ( id bigserial primary key, user_id varchar(36), origin_url varchar(255),deleted boolean default false, CONSTRAINT origin_url_unique UNIQUE (origin_url) )")

	return err
}

func (s *Database) Get(ctx context.Context, id int) (CreateURL, error) {
	var createURL CreateURL
	var deleted bool
	row := s.db.QueryRowContext(ctx, "SELECT id, user_id, origin_url, deleted FROM url WHERE id = $1", id)
	err := row.Scan(&createURL.ID, &createURL.User, &createURL.URL, &deleted)
	if err != nil {
		return CreateURL{}, err
	}

	if deleted {
		return CreateURL{}, ErrDeleted
	}

	return createURL, nil
}

func (s *Database) GetOriginURL(ctx context.Context, originURL string) (CreateURL, error) {
	var createURL CreateURL

	row := s.db.QueryRowContext(ctx, "SELECT id, user_id, origin_url FROM url WHERE origin_url = $1 AND deleted = false", originURL)
	err := row.Scan(&createURL.ID, &createURL.User, &createURL.URL)
	if err != nil {
		return CreateURL{}, err
	}

	return createURL, nil
}

func (s *Database) GetUser(ctx context.Context, userID string) ([]CreateURL, error) {
	rows := make([]CreateURL, 0)

	r, err := s.db.QueryContext(ctx, "SELECT id, user_id, origin_url FROM url WHERE user_id = $1 AND deleted = false", userID)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	for r.Next() {
		var createURL CreateURL
		err := r.Scan(&createURL.ID, &createURL.User, &createURL.URL)
		if err != nil {
			return nil, err
		}

		rows = append(rows, createURL)
	}

	err = r.Err()
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (s *Database) Set(ctx context.Context, createURL CreateURL) (int, error) {
	var id int

	sqlStatement := "INSERT INTO url (user_id, origin_url) VALUES ($1, $2) RETURNING id"
	err := s.db.QueryRowContext(ctx, sqlStatement, createURL.User, createURL.URL).Scan(&id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (s *Database) PutBatch(ctx context.Context, shortBatch []ShortenBatch) ([]ShortenBatch, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	sqlStatement := "INSERT INTO url (user_id, origin_url) VALUES ($1, $2) RETURNING id"
	stmt, err := tx.PrepareContext(ctx, sqlStatement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for id := range shortBatch {
		err = stmt.QueryRowContext(ctx, shortBatch[id].User, shortBatch[id].URL).Scan(&shortBatch[id].ID)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return shortBatch, nil
}

func (s *Database) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Database) DeleteBatch(ctx context.Context, ids []int) error {
	var strIds []string
	for _, id := range ids {
		strIds = append(strIds, fmt.Sprintf("%d", id))
	}

	if len(strIds) == 0 {
		return nil
	}

	stmt := fmt.Sprintf("UPDATE url set deleted = true WHERE id IN (%s)", strings.Join(strIds, ","))
	_, err := s.db.Exec(stmt)

	return err
}
