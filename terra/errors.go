package terra

import "fmt"

type InvalidEnvError struct {
	env string
}

func (e *InvalidEnvError) Error() string {
	return fmt.Sprintf("invalid %s environment", e.env)
}