package aws

import (
	"antelope/models"
	"antelope/models/types"
)

func ApiError(err error, msg string) (cError types.CustomCPError) {

	customError := types.CustomCPError{
		Error:       msg,
		Description: err.Error(),
		StatusCode:  int(models.CloudStatusCode),
	}

	return customError

}
