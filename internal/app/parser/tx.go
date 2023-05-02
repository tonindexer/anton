package parser

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

func (s *Service) parseDirectedMessage(ctx context.Context, acc *core.AccountData, message *core.Message, ret *core.MessagePayload) error {
	if acc == nil {
		return errors.Wrap(app.ErrImpossibleParsing, "no account data")
	}
	if len(acc.Types) == 0 {
		return errors.Wrap(app.ErrImpossibleParsing, "no interfaces")
	}

	operation, err := s.contractRepo.GetOperationByID(ctx, acc.Types, acc.Address == message.SrcAddress, message.OperationID)
	if errors.Is(err, core.ErrNotFound) {
		return errors.Wrap(app.ErrImpossibleParsing, "unknown operation")
	}
	if err != nil {
		return errors.Wrap(err, "get contract operations")
	}
	ret.OperationName = operation.Name

	// set src and dst contract types
	if acc.Address == message.SrcAddress {
		ret.SrcContract = operation.ContractName
	} else {
		ret.DstContract = operation.ContractName
	}

	msgParsed, err := operation.Schema.New()
	if err != nil {
		return errors.Wrapf(err, "creating struct from %s/%s schema", operation.ContractName, operation.Name)
	}

	payloadCell, err := cell.FromBOC(message.Body)
	if err != nil {
		return errors.Wrap(err, "msg body from boc")
	}
	payloadSlice := payloadCell.BeginParse()

	if err = tlb.LoadFromCell(msgParsed, payloadSlice); err != nil {
		return errors.Wrap(err, "load from cell")
	}

	ret.DataJSON, err = json.Marshal(msgParsed)
	if err != nil {
		return errors.Wrap(err, "json marshal parsed payload")
	}

	ret.MinterAddress = acc.MinterAddress

	return nil
}

func (s *Service) ParseMessagePayload(ctx context.Context, src, dst *core.AccountData, message *core.Message) (*core.MessagePayload, error) {
	var err = app.ErrImpossibleParsing // save message parsing error to a database to look at it later

	// you can parse separately incoming messages to known contracts and outgoing message from them

	ret := &core.MessagePayload{
		Type:        message.Type,
		Hash:        message.Hash,
		SrcAddress:  message.SrcAddress,
		DstAddress:  message.DstAddress,
		Amount:      message.Amount,
		BodyHash:    message.BodyHash,
		OperationID: message.OperationID,
		CreatedLT:   message.CreatedLT,
		CreatedAt:   message.CreatedAt,
	}
	if len(message.Body) == 0 {
		return nil, errors.Wrap(app.ErrImpossibleParsing, "no message body")
	}

	errIn := s.parseDirectedMessage(ctx, dst, message, ret)
	if errIn != nil && !errors.Is(errIn, app.ErrImpossibleParsing) {
		log.Warn().Err(errIn).
			Hex("tx_hash", message.SourceTxHash).
			Str("dst_addr", dst.Address.Base64()).
			Uint32("op_id", message.OperationID).Msgf("parse dst %v message", dst.Types)
		err = errors.Wrap(errIn, "incoming")
	}
	if errIn == nil {
		return ret, nil
	}

	errOut := s.parseDirectedMessage(ctx, src, message, ret)
	if errOut != nil && !errors.Is(errOut, app.ErrImpossibleParsing) {
		log.Warn().Err(errOut).
			Hex("tx_hash", message.SourceTxHash).
			Str("src_addr", src.Address.Base64()).
			Uint32("op_id", message.OperationID).Msgf("parse src %v message", src.Types)
		err = errors.Wrap(errOut, "outgoing")
	}
	if errOut == nil {
		return ret, nil
	}

	return ret, err
}
