package handlers

import (
	"fmt"
	"io"

	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
)

type URL struct {
	Url map[string]string
}

func QueryHandler(w http.ResponseWriter, r *http.Request) {
	// извлекаем фрагмент query= из URL запроса search?query=something
	// q := r.URL.Query().Get("query")
	// if q == "" {
	// 	http.Error(w, "The query parameter is missing", http.StatusBadRequest)
	// 	return
	// }
	ct := r.Header.Get("Content-Type")
	fmt.Println("ct:", ct)
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(requestDump))
	if r.Method == "GET" {
		vars := mux.Vars(r)
		id := vars["id"]
		fmt.Println(id)
		//if id == "qvery" {
		//	url := "http://localhost:8080/GET/" + id
		NewUrl := "https://practicum.yandex.ru/learn/go-developer/"
		uri := r.URL.Path
		// if uri == "/" {
		// 	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		// }
		fmt.Println(uri)
		w.Header().Set("content-type", "text/html; charset=UTF-8")
		w.Header().Set("Location", NewUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(NewUrl))
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	//}
	// в нашем случае q примет значение "something"
	// продолжаем обработку запроса

}

func BodyHandler(w http.ResponseWriter, r *http.Request) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(requestDump))
	if r.Method == "POST" {
		url := URL{}
		url.Url = make(map[string]string)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
		}
		url.Url["quwery"] = fmt.Sprintf("%v", body)
		//fmt.Println([]byte(body))
		// err = json.Unmarshal([]byte(body), &url.Url)
		// if err != nil {
		// 	log.Println("ERR", err)
		// }
		NewUrl := "http://localhost:8080/quvery"
		// uri := r.URL.Path
		// if uri == "/" {
		// 	http.Redirect(w, r, NewUrl, 201)
		// }
		w.Header().Set("Location", NewUrl)
		w.WriteHeader(http.StatusCreated)
		fmt.Println(url.Url)
		w.Write([]byte(NewUrl))

	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}
