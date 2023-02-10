package indexer

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/iam047801/tonidx/internal/core"
)

func (s *Service) parseMessagePayloads(ctx context.Context, messages []*core.Message, accountMap map[string]*core.AccountState) (ret []*core.MessagePayload) {
	for _, msg := range messages {
		if msg.Type != core.Internal {
			continue // TODO: external message parsing (?)
		}

		src, ok := accountMap[msg.SrcAddress]
		if !ok {
			log.Debug().Str("src_addr", msg.SrcAddress).Msg("cannot find src account")
			continue
		}
		dst, ok := accountMap[msg.DstAddress]
		if !ok {
			log.Debug().Str("src_addr", msg.SrcAddress).Msg("cannot find src account")
			continue
		}

		payload, err := s.parser.ParseMessagePayload(ctx, src, dst, msg)
		if errors.Is(err, core.ErrNotAvailable) {
			continue
		}
		if err != nil {
			log.Error().Err(err).Hex("msg_hash", msg.BodyHash).Hex("tx_hash", msg.TxHash).Msg("parse message payload")
			continue
		}
		ret = append(ret, payload)
	}

	return ret
}
