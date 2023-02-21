package reject

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

const (
	genericUnexpectedError string = "error.marketplace.generic.unexpected"
	cannotParseParams      string = "error.marketplace.generic.cannot-parse-params"
	cannotParseBody        string = "error.marketplace.generic.cannot-parse-payload"
	genericNotFound        string = "error.marketplace.generic.not-found"
)

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
