package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type Storage interface {
	Get(id int) (string, error)
	Set(data string) (int, error)
}

type Models struct {
	Model   map[int]string
	Counter int
	File    *os.File
	ticker  *time.Ticker
	done    chan bool
}

type CreateURL struct {
	ID  int
	URL string
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

func loadFile(file *os.File) (int, map[int]string, error) {
	var lastID int = 0
	var urls = make(map[int]string)
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

		urls[createURL.ID] = createURL.URL
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

func (md *Models) Get(id int) (string, error) {
	if url, ok := md.Model[id]; ok {
		return url, nil
	}

	return "", fmt.Errorf("id %d have not found", id)
}

func (md *Models) Set(url string) (int, error) {
	id := md.Counter
	md.Counter++

	md.Model[id] = url

	return id, nil
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

	for ID, URL := range md.Model {
		err := md.WriteCreateURLFile(CreateURL{
			ID:  ID,
			URL: URL,
		})

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
