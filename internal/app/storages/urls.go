package storages

type Record struct {
	ID   uint64
	User string
	URL  string
}

type BatchRecord struct {
	ID            uint64
	User          string
	URL           string
	CorrelationID string
}
