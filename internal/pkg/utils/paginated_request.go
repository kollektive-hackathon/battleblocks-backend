package utils

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
)

const (
	pageSizeMissing  string = "error.request.page-size-missing"
	pageTokenMissing string = "error.request.page-token-missing"
)

type PageRequest struct {
	Size   int
	Token  int
	Offset int
}

func NewPageRequest(c *gin.Context) (PageRequest, *reject.ProblemWithTrace) {
	pageSize, pageSizeError := strconv.Atoi(c.Query("page_size"))

	if pageSizeError != nil {
		return PageRequest{}, &reject.ProblemWithTrace{
			Problem: reject.NewProblem().
				WithTitle("Page size not specified").
				WithStatus(http.StatusBadRequest).
				WithCode(pageSizeMissing).
				Build(),
			Cause: pageSizeError,
		}
	}

	pageToken, pageTokenError := strconv.Atoi(c.Query("page_token"))

	if pageTokenError != nil {
		return PageRequest{}, &reject.ProblemWithTrace{
			Problem: reject.NewProblem().
				WithTitle("Page token not specified").
				WithStatus(http.StatusBadRequest).
				WithCode(pageTokenMissing).
				Build(),
			Cause: pageTokenError,
		}
	}

	var offset int
	if pageSize > 100 {
		offset = pageToken * 100
	} else {
		offset = pageSize * pageToken
	}

	return PageRequest{
		Size:   pageSize,
		Token:  pageToken,
		Offset: offset,
	}, nil
}
