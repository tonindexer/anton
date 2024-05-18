package parser

import (
	"encoding/base64"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/lru"
)

var _ app.ParserService = (*Service)(nil)

const itemsMinterCacheLen = 317750

type Service struct {
	*app.ParserConfig

	accountParseSemaphore chan struct{}

	itemsMinterCache *lru.Cache[addr.Address, addr.Address]

	bcConfigBase64 string
}

func NewService(cfg *app.ParserConfig) *Service {
	s := new(Service)
	s.ParserConfig = cfg
	s.bcConfigBase64 = base64.StdEncoding.EncodeToString(cfg.BlockchainConfig.ToBOC())
	s.accountParseSemaphore = make(chan struct{}, cfg.MaxAccountParsingWorkers)
	s.itemsMinterCache = lru.New[addr.Address, addr.Address](itemsMinterCacheLen)
	return s
}
