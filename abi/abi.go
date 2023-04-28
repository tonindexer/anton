package abi

type InterfaceDesc struct {
	InterfaceName string           `json:"interface_name"`
	InMessages    []*OperationDesc `json:"in_messages"`
	OutMessages   []*OperationDesc `json:"out_messages"`
	GetMethods    []*GetMethodDesc `json:"get_methods"`
}
