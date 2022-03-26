package storages

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type SimpleStorage struct {
	Start   uint64
	Records map[uint64]Record
	File    *os.File
	ticker  *time.Ticker
	done    chan bool
}

func CreateSimpleStorage(filename string, syncTime int) (*SimpleStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return nil, err
	}

	lastID, records, err := loadDataFromFile(file)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(time.Duration(syncTime) * time.Minute)
	done := make(chan bool)
	simpleStorage := &SimpleStorage{
		Start:   lastID + 1,
		Records: records,
		File:    file,
		ticker:  ticker,
		done:    done,
	}

	go simpleStorage.synchronize()

	return simpleStorage, nil
}

func loadDataFromFile(file *os.File) (uint64, map[uint64]Record, error) {
	var lastID uint64 = 0
	var urls = make(map[uint64]Record)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Bytes()

		record := Record{}
		err := json.Unmarshal(data, &record)
		if err != nil {
			return 0, nil, err
		}

		if record.ID > lastID {
			lastID = record.ID
		}

		urls[record.ID] = record
	}

	return lastID, urls, nil
}

func (s *SimpleStorage) synchronize() {
	for {
		select {
		case <-s.done:
			return
		case <-s.ticker.C:
			err := s.updateDataFile()
			if err != nil {
				return
			}
		}
	}
}

func (s *SimpleStorage) Get(ctx context.Context, id uint64) (Record, error) {
	if record, ok := s.Records[id]; ok {
		return record, nil
	}

	return Record{}, fmt.Errorf("id %d have not found", id)
}

func (s *SimpleStorage) GetByOriginURL(ctx context.Context, originURL string) (Record, error) {
	for _, record := range s.Records {
		if record.URL == originURL {
			return record, nil
		}
	}

	return Record{}, fmt.Errorf("originURL %s have not found", originURL)
}

func (s *SimpleStorage) GetByUser(ctx context.Context, userID string) ([]Record, error) {
	records := make([]Record, 0)

	for _, record := range s.Records {
		if record.User == userID {
			records = append(records, record)
		}
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("records with user_id %s have not found", userID)
	}

	return records, nil
}

func (s *SimpleStorage) Put(ctx context.Context, record Record) (uint64, error) {
	record.ID = s.Start
	s.Start++

	s.Records[record.ID] = record

	return record.ID, nil
}

func (s *SimpleStorage) Close() error {
	s.ticker.Stop()
	s.done <- true
	err := s.updateDataFile()

	if err != nil {
		return err
	}

	return s.File.Close()
}

func (s *SimpleStorage) updateDataFile() error {
	err := s.File.Truncate(0)
	if err != nil {
		return err
	}

	_, err = s.File.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	for _, record := range s.Records {
		err := s.writeRecordToFile(record)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SimpleStorage) writeRecordToFile(record Record) error {
	data, err := json.Marshal(record)

	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = s.File.Write(data)

	return err
}

func (s *SimpleStorage) PutRecords(ctx context.Context, records []BatchRecord) ([]BatchRecord, error) {
	return nil, fmt.Errorf("method has not implemented")
}

func (s *SimpleStorage) Ping(ctx context.Context) error {
	return nil
}
