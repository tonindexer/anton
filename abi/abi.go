package abi

import "github.com/xssnick/tonutils-go/address"

type ContractName string

type InterfaceDesc struct {
	InterfaceName ContractName             `json:"interface_name"`
	Addresses     []*address.Address       `json:"addresses,omitempty"`
	CodeBoc       string                   `json:"code_boc,omitempty"`
	Definitions   map[string]TLBFieldsDesc `json:"definitions,omitempty"`
	InMessages    []*OperationDesc         `json:"in_messages,omitempty"`
	OutMessages   []*OperationDesc         `json:"out_messages,omitempty"`
	GetMethods    []*GetMethodDesc         `json:"get_methods,omitempty"`
	ContractData  TLBFieldsDesc            `json:"contract_data,omitempty"`
}
