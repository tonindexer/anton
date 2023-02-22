package abi

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// https://github.com/TelegramMessenger/telemint/tree/main/func

type (
	TeleitemAuctionConfig struct {
		BeneficiaryAddress *address.Address `tlb:"addr"`
		InitialMinBid      tlb.Coins        `tlb:"."`
		MaxBid             tlb.Coins        `tlb:"."`
		MinBidStep         uint8            `tlb:"## 8"`
		MinExtendTime      uint32           `tlb:"## 32"`
		Duration           uint32           `tlb:"## 32"`
	}
	TelemintRoyaltyParams struct {
		Numerator   uint16           `tlb:"## 16"`
		Denominator uint16           `tlb:"## 16"`
		Destination *address.Address `tlb:"addr"`
	}
	TelemintTokenInfo struct {
		Name   *TelemintText `tlb:"."`
		Domain *TelemintText `tlb:"."`
	}

	TelemintMsgDeploy struct {
		_             tlb.Magic              `tlb:"#4637289b"`
		Sig           []byte                 `tlb:"bits 512"`
		SubwalletID   uint32                 `tlb:"## 32"`
		ValidSince    uint32                 `tlb:"## 32"`
		ValidTill     uint32                 `tlb:"## 32"`
		TokenName     *TelemintText          `tlb:"."`
		Content       *cell.Cell             `tlb:"^"`
		AuctionConfig *TeleitemAuctionConfig `tlb:"^"`
		RoyaltyParams *TelemintRoyaltyParams `tlb:"maybe ^"`
	}
	TeleitemMsgDeploy struct {
		_             tlb.Magic              `tlb:"#299a3e15"`
		SenderAddress *address.Address       `tlb:"addr"`
		Bid           tlb.Coins              `tlb:"."`
		Info          *TelemintTokenInfo     `tlb:"^"`
		Content       *cell.Cell             `tlb:"^"`
		AuctionConfig *TeleitemAuctionConfig `tlb:"^"`
		RoyaltyParams *TelemintRoyaltyParams `tlb:"^"`
	}
	TeleitemStartAuction struct {
		_             tlb.Magic              `tlb:"#487a8e81"`
		QueryID       uint64                 `tlb:"## 64"`
		AuctionConfig *TeleitemAuctionConfig `tlb:"^"`
	}
	TeleitemCancelAuction struct {
		_       tlb.Magic `tlb:"#371638ae"`
		QueryID uint64    `tlb:"## 64"`
	}
	TeleitemOK struct {
		_ tlb.Magic `tlb:"#a37a0983"`
	}
	TeleitemOutbidNotification struct {
		_ tlb.Magic `tlb:"#557cea20"`
	}
)
