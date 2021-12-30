package handlers

import (
	"fmt"
	"io"
	"strings"

	"net/http"
	"net/http/httputil"
)

type Url struct {
	urlmap map[string]string
	count  int
}

func (u *Url) AddUrl(url string) map[string]string {
	u.urlmap = map[string]string{}
	u.count = 1
	key := fmt.Sprintf("%d", u.count)
	u.urlmap[key] = url
	u.count++
	return u.urlmap
}

func NewUrl() *Url {
	return &Url{
		urlmap: make(map[string]string),
		count:  1,
	}
}

func (u *Url) QueryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(requestDump))
	if r.Method == "GET" {
		key := strings.TrimPrefix(r.URL.Path, "/")
		url := u.urlmap[key]
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println("invalid key", key)
			return
		}
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (u *Url) BodyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(requestDump))
	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
		}
		mapUrl := u.AddUrl(string(body))
		for key, val := range mapUrl {
			fmt.Println(val)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("http://localhost:8080/" + key))

		}

	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
