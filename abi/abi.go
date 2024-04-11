package abi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/tonindexer/anton/addr"
)

type ContractName string

type InterfaceDesc struct {
	Name         ContractName              `json:"interface_name"`
	Addresses    []*addr.Address           `json:"addresses,omitempty"`
	CodeBoc      string                    `json:"code_boc,omitempty"`
	Definitions  map[TLBType]TLBFieldsDesc `json:"definitions,omitempty"`
	InMessages   []OperationDesc           `json:"in_messages,omitempty"`
	OutMessages  []OperationDesc           `json:"out_messages,omitempty"`
	GetMethods   []GetMethodDesc           `json:"get_methods,omitempty"`
	ContractData TLBFieldsDesc             `json:"contract_data,omitempty"`
}

func RegisterDefinitions(definitions map[TLBType]TLBFieldsDesc, depth ...int) error {
	noDef := map[TLBType]TLBFieldsDesc{}
	for dn, d := range definitions {
		dt, err := tlbParseDesc(nil, d)
		if err != nil && strings.Contains(err.Error(), "cannot find definition") {
			noDef[dn] = d
			continue
		}
		if err != nil {
			return errors.Wrapf(err, "parse '%s' definition", dn)
		}

		if dt.Field(0).Type == typeNameMap[TLBTag] {
			// if the first struct field has tag,
			// we register it for the use in unions
			tlb.RegisterWithName(string(dn), reflect.New(dt).Elem().Interface())
		}

		registeredDefinitions[dn] = d
	}

	if len(noDef) == 0 {
		return nil
	}

	var currentDepth, maxDepth = 0, 16
	if len(depth) > 0 {
		currentDepth = depth[0]
	}
	if len(depth) > 1 {
		maxDepth = depth[1]
	}

	if currentDepth > maxDepth {
		var faultNames []string
		for dn := range noDef {
			faultNames = append(faultNames, string(dn))
		}
		return fmt.Errorf("cannot register [%s] definitions", strings.Join(faultNames, ", "))
	}

	return RegisterDefinitions(noDef, currentDepth+1, maxDepth)
}
