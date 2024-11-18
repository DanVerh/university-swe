package errorHandling

import (
	"log"
	"net/http"
)

type Error struct {
	w               http.ResponseWriter
	statusCode      int
	responseMessage string
	errorMessage    error
}

func (e *Error) Init(w http.ResponseWriter, statusCode int, responseMessage string, errorMessage error) {
	e.w = w
	e.statusCode = statusCode
	e.responseMessage = responseMessage
	e.errorMessage = errorMessage
}

func ThrowError(w http.ResponseWriter, statusCode int, responseMessage string, errorMessage error) {
	e := &Error{}
	e.Init(w, statusCode, responseMessage, errorMessage)

	if errorMessage == nil {
		log.Println(e.responseMessage)
	} else {
		log.Printf(e.responseMessage, "%v: ", e.errorMessage)
	}

	http.Error(w, e.responseMessage, e.statusCode)
}
