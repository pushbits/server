package oauth

const (
	// StatusError is the default generic error status
	StatusError ResponseStatus = "error"
	// StatusSuccess is the default generic success status
	StatusSuccess ResponseStatus = "success"
)

// ResponseStatus holds the status returned by a response struct
type ResponseStatus string

// JSONResponse holds a struct for displaying a message to a user in JSON format
type JSONResponse struct {
	Status  ResponseStatus `json:"status"`
	Message string         `json:"message"`
	Data    interface{}    `json:"data"`
}

// TODO cubicroot remove and use HTTP status codes instead like the rest of the application does
