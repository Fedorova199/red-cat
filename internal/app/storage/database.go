package storage

import (
	"context"
	"database/sql"
)

type Database struct {
	db *sql.DB
}

func CreateDatabase(db *sql.DB) (*Database, error) {
	database := &Database{
		db: db,
	}

	err := database.init()
	if err != nil {
		return nil, err
	}

	return database, nil
}

func (s *Database) init() error {
	_, err := s.db.Exec("CREATE TABLE IF NOT EXISTS url ( id bigserial primary key, user_id varchar(36), origin_url varchar(255) )")

	return err
}

func (s *Database) Get(ctx context.Context, id int) (CreateURL, error) {
	var createURL CreateURL

	row := s.db.QueryRowContext(ctx, "SELECT id, user_id, origin_url FROM url WHERE id = $1", id)
	err := row.Scan(&createURL.ID, &createURL.User, &createURL.URL)
	if err != nil {
		return CreateURL{}, err
	}

	return createURL, nil
}

func (s *Database) GetByUser(ctx context.Context, userID string) ([]CreateURL, error) {
	records := make([]CreateURL, 0)

	rows, err := s.db.QueryContext(ctx, "SELECT id, user_id, origin_url FROM url WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var createURL CreateURL
		err := rows.Scan(&createURL.ID, &createURL.User, &createURL.URL)
		if err != nil {
			return nil, err
		}

		records = append(records, createURL)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (s *Database) Set(ctx context.Context, createURL CreateURL) (int, error) {
	var id int

	sqlStatement := "INSERT INTO url (user_id, origin_url) VALUES ($1, $2) RETURNING id"
	err := s.db.QueryRowContext(ctx, sqlStatement, createURL.User, createURL.URL).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}
