package validator

type createValidator struct{}

func ForCreate() *createValidator {
	return &createValidator{}
}
