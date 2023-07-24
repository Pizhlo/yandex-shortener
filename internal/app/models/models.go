package models

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type RequestAPI struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

type ResponseAPI struct {
	ID       string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}
