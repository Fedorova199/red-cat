package handlers

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Fedorova199/red-cat/internal/app/middlewares"
	"github.com/Fedorova199/red-cat/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_POSTHandler(t *testing.T) {
	file, err := ioutil.TempFile("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	type want struct {
		contentType string
		statusCode  int
		id          string
	}
	tests := []struct {
		name    string
		path    string
		body    string
		storage storage.Storage
		want    want
	}{
		{
			name: "simple test #1",
			storage: &storage.Models{
				Model: map[int]storage.CreateURL{
					1: {
						ID:   1,
						User: "user",
						URL:  "test1.ru",
					},
					2: {
						ID:   2,
						User: "user",
						URL:  "test2.ru",
					},
				},
				Counter: 3,
				File:    file,
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  201,
				id:          "3",
			},
			path: "/",
			body: "test1.ru",
		},
		{
			name: "empty body #2",
			storage: &storage.Models{
				Model: map[int]storage.CreateURL{
					1: {
						ID:   1,
						User: "user",
						URL:  "test1.ru",
					},
					2: {
						ID:   2,
						User: "user",
						URL:  "test2.ru",
					},
				},
				File: file,
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  201,
				id:          "0",
			},
			path: "/",
			body: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.storage, "test.ru", []Middleware{
				middlewares.GzipHandle{},
				middlewares.UngzipHandle{},
				middlewares.NewAuthenticator([]byte("secret key")),
			})
			ts := httptest.NewServer(handler)
			defer ts.Close()

			resp, body := testRequest(t, ts, http.MethodPost, tt.path, strings.NewReader(tt.body))
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Contains(t, body, tt.want.id)
		})
	}
}

func TestHandler_GETHandler(t *testing.T) {
	file, err := ioutil.TempFile("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	type want struct {
		contentType string
		statusCode  int
		redirectURL string
	}
	tests := []struct {
		name    string
		path    string
		storage storage.Storage
		want    want
	}{
		{
			name: "simple test #1",
			storage: &storage.Models{
				Model: map[int]storage.CreateURL{
					1: {
						ID:   1,
						User: "user",
						URL:  "test1.ru",
					},
					2: {
						ID:   2,
						User: "user",
						URL:  "test2.ru",
					},
				},
				Counter: 3,
				File:    file,
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  307,
				redirectURL: "test2.ru",
			},
			path: "/2",
		},
		{
			name: "wrong id #2",
			storage: &storage.Models{
				Model: map[int]storage.CreateURL{
					1: {
						ID:   1,
						User: "user",
						URL:  "test1.ru",
					},
					2: {
						ID:   2,
						User: "user",
						URL:  "test2.ru",
					},
				},
				Counter: 3,
				File:    file,
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  500,
				redirectURL: "",
			},
			path: "/1009",
		},
		{
			name: "empty id #3",
			storage: &storage.Models{
				Model: map[int]storage.CreateURL{
					1: {
						ID:   1,
						User: "user",
						URL:  "test1.ru",
					},
					2: {
						ID:   2,
						User: "user",
						URL:  "test2.ru",
					},
				},
				Counter: 3,
				File:    file,
			},
			want: want{
				contentType: "",
				statusCode:  405,
				redirectURL: "",
			},
			path: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.storage, "test.ru", []Middleware{
				middlewares.GzipHandle{},
				middlewares.UngzipHandle{},
				middlewares.NewAuthenticator([]byte("secret key")),
			})
			ts := httptest.NewServer(handler)
			defer ts.Close()

			resp, _ := testRequest(t, ts, http.MethodGet, tt.path, strings.NewReader(""))
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			//assert.Equal(t, tt.want.redirectURL, resp.Header.Get("Location"))
		})
	}
}

func TestHandler_JSONHandler(t *testing.T) {
	file, err := ioutil.TempFile("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	type want struct {
		contentType string
		statusCode  int
		id          string
	}
	tests := []struct {
		name    string
		path    string
		body    string
		storage storage.Storage
		want    want
	}{
		{
			name: "simple test #1",
			storage: &storage.Models{
				Model: map[int]storage.CreateURL{
					1: {
						ID:   1,
						User: "user",
						URL:  "test1.ru",
					},
					2: {
						ID:   2,
						User: "user",
						URL:  "test2.ru",
					},
				},
				Counter: 3,
				File:    file,
			},
			want: want{
				contentType: "application/json",
				statusCode:  201,
				id:          "3",
			},
			path: "/api/shorten",
			body: "{\"url\": \"test1.ru\"}",
		},
		{
			name: "empty json #2",
			storage: &storage.Models{
				Model: map[int]storage.CreateURL{
					1: {
						ID:   1,
						User: "user",
						URL:  "test1.ru",
					},
					2: {
						ID:   2,
						User: "user",
						URL:  "test2.ru",
					},
				},
				Counter: 3,
				File:    file,
			},
			want: want{
				contentType: "application/json",
				statusCode:  201,
				id:          "3",
			},
			path: "/api/shorten",
			body: "{}",
		},
		{
			name: "wrong json #3",
			storage: &storage.Models{
				Model: map[int]storage.CreateURL{
					1: {
						ID:   1,
						User: "user",
						URL:  "test1.ru",
					},
					2: {
						ID:   2,
						User: "user",
						URL:  "test2.ru",
					},
				},
				Counter: 3,
				File:    file,
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  500,
				id:          "",
			},
			path: "/api/shorten",
			body: "{",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.storage, "test.ru", []Middleware{
				middlewares.GzipHandle{},
				middlewares.UngzipHandle{},
				middlewares.NewAuthenticator([]byte("secret key")),
			})
			ts := httptest.NewServer(handler)
			defer ts.Close()

			resp, body := testRequest(t, ts, http.MethodPost, tt.path, strings.NewReader(tt.body))
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Contains(t, body, tt.want.id)
		})
	}
}

func TestNewHandler(t *testing.T) {
	storage := &storage.Models{
		Counter: 3,
		Model: map[int]storage.CreateURL{
			1: {
				ID:   1,
				User: "user",
				URL:  "test1.ru",
			},
			2: {
				ID:   2,
				User: "user",
				URL:  "test2.ru",
			},
		},
	}
	handler := NewHandler(storage, "test.ru", []Middleware{
		middlewares.GzipHandle{},
		middlewares.UngzipHandle{},
	})
	assert.Implements(t, (*http.Handler)(nil), handler)
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}
