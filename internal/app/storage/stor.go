package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type Models struct {
	Counter int
	Model   map[int]CreateURL
	File    *os.File
	ticker  *time.Ticker
	done    chan bool
}

func NewModels(filename string, syncTime int) (*Models, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return nil, err
	}

	lastID, model, err := CreateDataFile(file)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(time.Duration(syncTime) * time.Minute)
	done := make(chan bool)
	simpleStorage := &Models{
		Counter: lastID + 1,
		Model:   model,
		File:    file,
		ticker:  ticker,
		done:    done,
	}

	go simpleStorage.synchronize()

	return simpleStorage, nil
}

func CreateDataFile(file *os.File) (int, map[int]CreateURL, error) {
	lastID := 0
	var urls = make(map[int]CreateURL)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Bytes()

		createURL := CreateURL{}
		err := json.Unmarshal(data, &createURL)
		if err != nil {
			return 0, nil, err
		}

		if createURL.ID > lastID {
			lastID = createURL.ID
		}

		urls[createURL.ID] = createURL
	}

	return lastID, urls, nil
}

func (md *Models) synchronize() {
	for {
		select {
		case <-md.done:
			return
		case <-md.ticker.C:
			err := md.updateDataFile()
			if err != nil {
				return
			}
		}
	}
}

func (md *Models) Get(ctx context.Context, id int) (CreateURL, error) {
	if createURL, ok := md.Model[id]; ok {
		return createURL, nil
	}

	return CreateURL{}, fmt.Errorf("id %d have not found", id)
}

func (md *Models) GetOriginURL(ctx context.Context, originURL string) (CreateURL, error) {
	for _, createURL := range md.Model {
		if createURL.URL == originURL {
			return createURL, nil
		}
	}

	return CreateURL{}, fmt.Errorf("originURL %s have not found", originURL)
}

func (md *Models) GetUser(ctx context.Context, userID string) ([]CreateURL, error) {
	model := make([]CreateURL, 0)

	for _, value := range md.Model {
		if value.User == userID {
			model = append(model, value)
		}
	}

	if len(model) == 0 {
		return nil, fmt.Errorf("model with user_id %s have not found", userID)
	}

	return model, nil
}

func (md *Models) Set(ctx context.Context, createURL CreateURL) (int, error) {
	createURL.ID = md.Counter
	md.Counter++

	md.Model[createURL.ID] = createURL

	return createURL.ID, nil
}

func (md *Models) Close() error {
	md.ticker.Stop()
	md.done <- true
	err := md.updateDataFile()

	if err != nil {
		return err
	}

	return md.File.Close()
}

func (md *Models) updateDataFile() error {
	err := md.File.Truncate(0)
	if err != nil {
		return err
	}

	_, err = md.File.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	for _, createURL := range md.Model {
		err := md.writeToFile(createURL)

		if err != nil {
			return err
		}
	}

	return nil
}

func (md *Models) writeToFile(record CreateURL) error {
	data, err := json.Marshal(record)

	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = md.File.Write(data)

	return err
}

func (md *Models) PutBatch(ctx context.Context, shortBatch []ShortenBatch) ([]ShortenBatch, error) {
	return nil, fmt.Errorf("method has not implemented")
}

func (md *Models) Ping(ctx context.Context) error {
	return nil
}
