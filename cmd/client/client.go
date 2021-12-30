package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func ClientGet() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8080/1", nil)
	if err != nil {
		fmt.Println(err)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer res.Body.Close()
	fmt.Println("BODY", string(body[0:1]))
	fmt.Println("Header", res.Header)
	// печатаем код ответа
	fmt.Println("Статус-кодGET ", res.StatusCode)

}

func ClientPost() {
	client := &http.Client{}
	var b = []byte("https://practicum.yandex.ru/")
	request, err := http.NewRequest("POST", "http://localhost:8080/", bytes.NewBuffer(b))
	if err != nil {
		fmt.Println(err)
	}
	request.Header.Set("Content-Type", "text/plain; charset=utf-8")
	client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if len(via) >= 2 {
			return errors.New("Остановлено после двух Redirect")
		}
		return nil
	}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} //////////////////////////
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(body))
	fmt.Println(response.Header)
	// печатаем код ответа
	fmt.Println("Статус-кодPOST ", response.StatusCode)
	defer response.Body.Close()
}

func main() {
	ClientPost()
	ClientGet()
}
