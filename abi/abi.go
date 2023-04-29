package abi

import (
	"reflect"

	"github.com/pkg/errors"

	"github.com/tonindexer/anton/addr"
)

type ContractName string

type InterfaceDesc struct {
	Name         ContractName             `json:"interface_name"`
	Addresses    []*addr.Address          `json:"addresses,omitempty"`
	CodeBoc      string                   `json:"code_boc,omitempty"`
	Definitions  map[string]TLBFieldsDesc `json:"definitions,omitempty"`
	InMessages   []OperationDesc          `json:"in_messages,omitempty"`
	OutMessages  []OperationDesc          `json:"out_messages,omitempty"`
	GetMethods   []GetMethodDesc          `json:"get_methods,omitempty"`
	ContractData TLBFieldsDesc            `json:"contract_data,omitempty"`
}

func (i *InterfaceDesc) RegisterDefinitions() error {
	for dn, d := range i.Definitions {
		v, err := d.New()
		if err != nil {
			return errors.Wrapf(err, "parse '%s' definition", dn)
		}
		t := reflect.TypeOf(v)
		typeNameMap[dn] = t
		typeNameRMap[t] = dn
	}
	return nil
}
