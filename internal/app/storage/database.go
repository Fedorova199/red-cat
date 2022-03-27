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

func (s *Database) GetByOriginURL(ctx context.Context, originURL string) (CreateURL, error) {
	var model CreateURL

	row := s.db.QueryRowContext(ctx, "SELECT id, user_id, origin_url FROM url WHERE origin_url = $1", originURL)
	err := row.Scan(&model.ID, &model.URL, &model.User)
	if err != nil {
		return CreateURL{}, err
	}

	return model, nil
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

func (s *Database) APIShortenBatch(ctx context.Context, models []ShortenBatch) ([]ShortenBatch, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	sqlStatement := "INSERT INTO url (user_id, origin_url) VALUES ($1, $2) RETURNING id"
	stmt, err := tx.PrepareContext(ctx, sqlStatement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for id := range models {
		err = stmt.QueryRowContext(ctx, models[id].User, models[id].URL).Scan(&models[id].ID)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return models, nil
}

func (s *Database) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
