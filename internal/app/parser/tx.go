package parser

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/iam047801/tonidx/internal/core"
)

func (s *Service) parseDirectedMessage(ctx context.Context, acc *core.AccountState, message *core.Message, ret *core.MessagePayload) error {
	if len(acc.Types) == 0 {
		return errors.Wrap(core.ErrNotAvailable, "no interfaces")
	}

	operation, err := s.abiRepo.GetContractOperationByID(ctx, acc, acc.Address == message.SrcAddress, message.OperationID)
	if errors.Is(err, core.ErrNotFound) {
		return errors.Wrap(core.ErrNotAvailable, "unknown operation")
	}
	if err != nil {
		return errors.Wrap(err, "get contract operations")
	}
	ret.OperationName = operation.Name

	if acc.Address == message.SrcAddress {
		ret.SrcContract = operation.ContractName
	} else {
		ret.DstContract = operation.ContractName
	}

	payloadCell, err := cell.FromBOC(message.Body)
	if err != nil {
		return errors.Wrap(err, "msg body from boc")
	}
	payloadSlice := payloadCell.BeginParse()

	parsed := reflect.New(reflect.StructOf(operation.StructSchema)).Interface()
	if err = tlb.LoadFromCell(parsed, payloadSlice); err != nil {
		// return errors.Wrapf(core.ErrNotAvailable, "load from cell (%s)", err.Error())
		return errors.Wrap(err, "load from cell")
	}

	parsedJSON, err := json.Marshal(parsed)
	if err != nil {
		return errors.Wrap(err, "json marshal parsed payload")
	}
	ret.DataJSON = string(parsedJSON)

	return nil
}

func (s *Service) ParseMessagePayload(ctx context.Context, src, dst *core.AccountState, message *core.Message) (*core.MessagePayload, error) {
	// you can parse separately incoming messages to known contracts and outgoing message from them

	ret := &core.MessagePayload{
		TxHash:      message.TxHash,
		BodyHash:    message.BodyHash,
		SrcAddress:  message.SrcAddress,
		DstAddress:  message.DstAddress,
		OperationID: message.OperationID,
	}
	if len(message.Body) == 0 {
		return nil, errors.Wrap(core.ErrNotAvailable, "no message body")
	}

	err := s.parseDirectedMessage(ctx, dst, message, ret)
	if err != nil && !errors.Is(err, core.ErrNotAvailable) {
		log.Warn().
			Err(err).
			Hex("tx_hash", message.TxHash).
			Str("dst_addr", dst.Address).
			Strs("dst_types", dst.Types).
			Uint32("op_id", message.OperationID).Msg("parse dst message")
	}
	if err == nil {
		return ret, nil
	}

	err = s.parseDirectedMessage(ctx, src, message, ret)
	if err != nil && !errors.Is(err, core.ErrNotAvailable) {
		log.Warn().
			Err(err).
			Hex("tx_hash", message.TxHash).
			Str("src_addr", src.Address).
			Strs("src_types", src.Types).
			Uint32("op_id", message.OperationID).Msg("parse src message")
	}
	if err == nil {
		return ret, nil
	}

	return nil, err
}
