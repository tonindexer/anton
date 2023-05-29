package parser

import (
	"sync"

	"github.com/tonindexer/anton/internal/app"
)

var _ app.ParserService = (*Service)(nil)

type Service struct {
	*app.ParserConfig

	emulatorMx sync.RWMutex
}

func NewService(cfg *app.ParserConfig) *Service {
	s := new(Service)
	s.ParserConfig = cfg
	return s
}
