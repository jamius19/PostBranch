package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/jamius19/postbranch/logger"
)

var log = logger.Logger
var validate *validator.Validate

func init() {
	validate = validator.New()
	log.Info("Initialized validator")

	//err := validate.RegisterValidation("initCon", repo.InitValidation)
	//if err != nil {
	//	log.Fatalf("Failed to register custom validation function: %s", err)
	//}
}

func Validate(val any) error {
	return validate.Struct(val)
}
