package controller

import (
	"github.com/gin-gonic/gin"
)

// Group is used to add a prefix to a list of Controllers.
type Group struct {
	controllers []Controller
	prefix      string
}

// This piece of code is to ensure that a type implements a certain interface at compile time.
// More info: https://github.com/uber-go/guide/blob/master/style.md#verify-interface-compliance
var _ Controller = (*Group)(nil)

func (g Group) PathPrefix() string {
	return g.prefix
}

func (g Group) RegisterHandlers(e *gin.RouterGroup) {
	for _, controller := range g.controllers {
		controller.RegisterHandlers(e.Group(controller.PathPrefix()))
	}
}

func NewGroup(prefix string, controllers ...Controller) Controller {
	return Group{
		controllers: controllers,
		prefix:      prefix,
	}
}
