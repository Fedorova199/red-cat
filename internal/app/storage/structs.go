package storage

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type ShortURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginURL     string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type CreateURL struct {
	ID   int
	User string
	URL  string
}

type ShortenBatch struct {
	ID            int
	User          string
	URL           string
	CorrelationID string
}
