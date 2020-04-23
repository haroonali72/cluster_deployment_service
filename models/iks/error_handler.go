package iks

import (
	"antelope/models/types"
)

func ApiError(err error, msg string, statusCode int) (cError types.CustomCPError) {

	customError := types.CustomCPError{
		Error:     msg,
		Description: err.Error(),
		StatusCode:  statusCode,
	}
	return customError

}
