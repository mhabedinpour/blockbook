package block

import (
	"blockbook/pkg/bcparser"
	"blockbook/pkg/controller"

	"github.com/gin-gonic/gin"
)

type Block struct {
	parser bcparser.Parser
}

// This piece of code is to ensure that a type implements a certain interface at compile time.
// More info: https://github.com/uber-go/guide/blob/master/style.md#verify-interface-compliance
var _ controller.Controller = (*Block)(nil)

func (b *Block) PathPrefix() string {
	return "/block"
}

func (b *Block) RegisterHandlers(engine *gin.RouterGroup) {
	engine.GET("/current", b.current)
}

func (b *Block) current(c *gin.Context) {
	controller.WriteSuccess(gin.H{
		"lastIndexedBlock": b.parser.CurrentBlockNumber(),
	}, c)
}

func New(parser bcparser.Parser) *Block {
	return &Block{
		parser: parser,
	}
}
