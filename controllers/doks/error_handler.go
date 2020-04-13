package doks

import (
	"antelope/models/doks"
	"errors"
	"github.com/asaskevich/govalidator"
)

func validateStruct (cluster doks.KubernetesCluster, token string ) error {

	_, err := govalidator.ValidateStruct(cluster)
	if err != nil {
		return errors.New("Json Structure Error: "+err.Error())
	}

	if token == "" {
		return errors.New("Token Is Empty")
	}

	return nil
}


