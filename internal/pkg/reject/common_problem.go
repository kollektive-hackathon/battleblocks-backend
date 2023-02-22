package reject

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

const (
	genericUnexpectedError string = "error.generic.unexpected"
	cannotParseParams      string = "error.generic.cannot-parse-params"
	invalidRequest         string = "error.generic.invalid-request-payload"
	cannotParseBody        string = "error.generic.cannot-parse-payload"
	genericNotFound        string = "error.generic.not-found"
)

func RequestValidationProblem() Problem {
	return NewProblem().
		WithTitle("Invalid request payload").
		WithStatus(http.StatusBadRequest).
		WithCode(invalidRequest).
		Build()
}

func RequestParamsProblem() Problem {
	return NewProblem().
		WithTitle("Invalid request parameters").
		WithStatus(http.StatusBadRequest).
		WithCode(cannotParseParams).
		Build()
}

func BodyParseProblem() Problem {
	return NewProblem().
		WithTitle("Cannot read payload").
		WithStatus(http.StatusBadRequest).
		WithCode(cannotParseBody).
		Build()
}

func NotFoundProblem() Problem {
	return NewProblem().
		WithTitle("Record not found").
		WithStatus(http.StatusNotFound).
		WithCode(genericNotFound).
		Build()
}

func UnexpectedProblem(err error) Problem {
	log.Warn().Err(err).Msg("Unexpected error while handling request: " + err.Error())
	return NewProblem().
		WithTitle("Unexpected error").
		WithStatus(http.StatusInternalServerError).
		WithCode(genericUnexpectedError).
		Build()
}
