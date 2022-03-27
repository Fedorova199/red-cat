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

type Storage interface {
	Get(tx context.Context, id int) (CreateURL, error)
	Set(tx context.Context, model CreateURL) (int, error)
	GetByUser(tx context.Context, userID string) ([]CreateURL, error)
}

type Models struct {
	Model   map[int]CreateURL
	Counter int
	File    *os.File
	ticker  *time.Ticker
	done    chan bool
}

type CreateURL struct {
	ID   int
	URL  string
	User string
}

func NewModels(filename string, syncTime int) (*Models, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return nil, err
	}

	lastID, urls, err := loadFile(file)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(time.Duration(syncTime) * time.Minute)
	done := make(chan bool)
	modelStor := &Models{
		Model:   urls,
		Counter: lastID + 1,
		File:    file,
		ticker:  ticker,
		done:    done,
	}

	go modelStor.synchronize()

	return modelStor, nil
}

func loadFile(file *os.File) (int, map[int]CreateURL, error) {
	var lastID = 0
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
			err := md.UpdateFile()
			if err != nil {
				return
			}
		}
	}
}

func (md *Models) Get(tx context.Context, id int) (CreateURL, error) {
	if model, ok := md.Model[id]; ok {
		return model, nil
	}

	return CreateURL{}, fmt.Errorf("id %d have not found", id)
}

func (md *Models) Set(tx context.Context, model CreateURL) (int, error) {
	model.ID = md.Counter
	md.Counter++

	md.Model[model.ID] = model

	return model.ID, nil
}

func (md *Models) GetByUser(tx context.Context, userID string) ([]CreateURL, error) {
	arrUsers := make([]CreateURL, 0)

	for _, val := range md.Model {
		if val.User == userID {
			arrUsers = append(arrUsers, val)
		}
	}
	if len(arrUsers) == 0 {
		return nil, fmt.Errorf("arrUsers with user_id %s have not found", userID)
	}
	return arrUsers, nil
}

func (md *Models) Close() error {
	md.ticker.Stop()
	md.done <- true
	err := md.UpdateFile()

	if err != nil {
		return err
	}

	return md.File.Close()
}

func (md *Models) UpdateFile() error {
	err := md.File.Truncate(0)
	if err != nil {
		return err
	}

	_, err = md.File.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	for _, val := range md.Model {
		err := md.WriteCreateURLFile(val)

		if err != nil {
			return err
		}
	}

	return nil
}

func (md *Models) WriteCreateURLFile(createURL CreateURL) error {
	data, err := json.Marshal(createURL)

	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = md.File.Write(data)

	return err
}
