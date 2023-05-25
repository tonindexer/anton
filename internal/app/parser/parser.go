package parser

import (
	"github.com/tonindexer/anton/internal/app"
)

var _ app.ParserService = (*Service)(nil)

type Service struct {
	*app.ParserConfig
}

func NewService(cfg *app.ParserConfig) *Service {
	s := new(Service)
	s.ParserConfig = cfg
	return s
}
