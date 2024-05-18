package parser

import (
	"encoding/base64"

	"github.com/tonindexer/anton/internal/app"
)

var _ app.ParserService = (*Service)(nil)

type Service struct {
	*app.ParserConfig

	accountParseSemaphore chan struct{}

	bcConfigBase64 string
}

func NewService(cfg *app.ParserConfig) *Service {
	s := new(Service)
	s.ParserConfig = cfg
	s.bcConfigBase64 = base64.StdEncoding.EncodeToString(cfg.BlockchainConfig.ToBOC())
	s.accountParseSemaphore = make(chan struct{}, cfg.MaxAccountParsingWorkers)
	return s
}
