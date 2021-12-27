package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	client := &http.Client{}
	var body = []byte("https://practicum.yandex.ru/learn/go-developer/")
	request, err := http.NewRequest("POST", "http://localhost:8080/POST/", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
	}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} //////////////////////////

	////////////////
	req, err := http.NewRequest("GET", "http://localhost:8080/GET/qvery", nil)
	if err != nil {
		fmt.Println(err)
	}
	///////////////////

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	// client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
	// 	if len(via) >= 2 {
	// 		return errors.New("остановлено после двух Redirect")
	// 	}
	// 	return nil
	// }

	// client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
	// 	if len(via) >= 2 {
	// 		return errors.New("остановлено после двух Redirect")
	// 	}
	// 	return nil
	// }
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer res.Body.Close()
	bod, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(bod))
	fmt.Println(res.Header)
	// печатаем код ответа
	fmt.Println("Статус-кодPOST ", response.StatusCode)
	fmt.Println("Статус-кодGET ", res.StatusCode)
	defer response.Body.Close()

	// читаем поток из тела ответа
	b, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// и печатаем его
	fmt.Println(string(b))
	fmt.Println(response.Header)

}
