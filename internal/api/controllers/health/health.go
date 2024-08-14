package health

import (
	"blockbook/pkg/controller"

	"github.com/gin-gonic/gin"
)

type Health struct {
}

// This piece of code is to ensure that a type implements a certain interface at compile time.
// More info: https://github.com/uber-go/guide/blob/master/style.md#verify-interface-compliance
var _ controller.Controller = (*Health)(nil)

func (h *Health) PathPrefix() string {
	return "/-"
}

func (h *Health) RegisterHandlers(engine *gin.RouterGroup) {
	engine.GET("/ready", h.ready)
	engine.GET("/live", h.live)
}

func (h *Health) ready(c *gin.Context) {
	controller.WriteSuccess(gin.H{}, c)
}

func (h *Health) live(c *gin.Context) {
	controller.WriteSuccess(gin.H{}, c)
	// TODO: check if we have the latest block information
}

func New() *Health {
	return &Health{}
}
