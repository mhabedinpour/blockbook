package address

import (
	"blockbook/pkg/bcparser"
	"blockbook/pkg/controller"

	"github.com/gin-gonic/gin"
)

type Address struct {
	parser bcparser.Parser
}

// This piece of code is to ensure that a type implements a certain interface at compile time.
// More info: https://github.com/uber-go/guide/blob/master/style.md#verify-interface-compliance
var _ controller.Controller = (*Address)(nil)

func (a *Address) PathPrefix() string {
	return "/address"
}

func (a *Address) RegisterHandlers(engine *gin.RouterGroup) {
	engine.GET("/:address/transactions", a.transactions)
	engine.POST("/subscribe", a.subscribe)
	engine.DELETE("/unsubscribe", a.unsubscribe)
}

func (a *Address) subscribe(c *gin.Context) {
	model, err := controller.BindBody[AddressModel](c)
	if err != nil {
		controller.WriteError(err, c)

		return
	}

	ok := a.parser.Subscribe(model.Address)
	if !ok {
		controller.WriteError(ErrAddressAlreadySubscribed, c)

		return
	}

	controller.WriteSuccess(gin.H{}, c)
}

func (a *Address) unsubscribe(c *gin.Context) {
	model, err := controller.BindBody[AddressModel](c)
	if err != nil {
		controller.WriteError(err, c)

		return
	}

	ok := a.parser.Unsubscribe(model.Address)
	if !ok {
		controller.WriteError(ErrAddressNotSubscribed, c)

		return
	}

	controller.WriteSuccess(gin.H{}, c)
}

func (a *Address) transactions(c *gin.Context) {
	model, err := controller.BindUri[AddressModel](c)
	if err != nil {
		controller.WriteError(err, c)

		return
	}

	txs := a.parser.Transactions(model.Address)
	if txs == nil {
		controller.WriteError(ErrAddressNotSubscribed, c)

		return
	}

	controller.WriteSuccess(gin.H{
		"transactions": txs,
	}, c)
}

func New(parser bcparser.Parser) *Address {
	return &Address{
		parser: parser,
	}
}
