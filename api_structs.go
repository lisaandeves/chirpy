package main

type responseError struct {
	Error string `json:"error"`
}

type userParams struct {
	Email string `json:"email"`
}

type userResponse struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
}

type chirpParams struct {
	Body   string `json:"body"`
	UserId string `json:"user_id"`
}

type chirpResponse struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body      string `json:"body"`
	UserId    string `json:"user_id"`
}
