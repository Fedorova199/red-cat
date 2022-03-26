package storages

import (
	"context"
	"database/sql"
)

type DatabaseStorage struct {
	db *sql.DB
}

func CreateDatabaseStorage(db *sql.DB) (*DatabaseStorage, error) {
	databaseStorage := &DatabaseStorage{
		db: db,
	}

	err := databaseStorage.init()
	if err != nil {
		return nil, err
	}

	return databaseStorage, nil
}

func (s *DatabaseStorage) init() error {
	_, err := s.db.Exec("CREATE TABLE IF NOT EXISTS url ( id bigserial primary key, user_id varchar(36), origin_url varchar(255), CONSTRAINT origin_url_unique UNIQUE (origin_url) )")

	return err
}

func (s *DatabaseStorage) Get(ctx context.Context, id uint64) (Record, error) {
	var record Record

	row := s.db.QueryRowContext(ctx, "SELECT id, user_id, origin_url FROM url WHERE id = $1", id)
	err := row.Scan(&record.ID, &record.User, &record.URL)
	if err != nil {
		return Record{}, err
	}

	return record, nil
}

func (s *DatabaseStorage) GetByOriginURL(ctx context.Context, originURL string) (Record, error) {
	var record Record

	row := s.db.QueryRowContext(ctx, "SELECT id, user_id, origin_url FROM url WHERE origin_url = $1", originURL)
	err := row.Scan(&record.ID, &record.User, &record.URL)
	if err != nil {
		return Record{}, err
	}

	return record, nil
}

func (s *DatabaseStorage) GetByUser(ctx context.Context, userID string) ([]Record, error) {
	records := make([]Record, 0)

	rows, err := s.db.QueryContext(ctx, "SELECT id, user_id, origin_url FROM url WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var record Record
		err := rows.Scan(&record.ID, &record.User, &record.URL)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (s *DatabaseStorage) Put(ctx context.Context, record Record) (uint64, error) {
	var id uint64

	sqlStatement := "INSERT INTO url (user_id, origin_url) VALUES ($1, $2) RETURNING id"
	err := s.db.QueryRowContext(ctx, sqlStatement, record.User, record.URL).Scan(&id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (s *DatabaseStorage) PutRecords(ctx context.Context, records []BatchRecord) ([]BatchRecord, error) {
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

	for id := range records {
		err = stmt.QueryRowContext(ctx, records[id].User, records[id].URL).Scan(&records[id].ID)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return records, nil
}

func (s *DatabaseStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
