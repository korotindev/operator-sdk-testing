package controller

import (
	"github.com/soxat/operator-sdk-testing/pkg/controller/application"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, application.Add)
}
