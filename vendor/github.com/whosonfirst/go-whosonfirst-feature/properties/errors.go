package properties

import (
	_ "errors"
	"fmt"
)

type MissingPropertyError struct {
	property string
}

func (e *MissingPropertyError) Error() string {
	return fmt.Sprintf("Missing '%s' property", e.property)
}

func (e *MissingPropertyError) String() string {
	return e.Error()
}

func MissingProperty(prop string) error {
	return &MissingPropertyError{property: prop}
}
