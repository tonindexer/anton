package abi

type FuncValueDesc struct {
	Name     string `json:"name"`
	GoType   string `json:"go_type"`
	FuncType string `json:"func_type"`
}

type GetMethodDesc struct {
	Name         string           `json:"name"`
	Arguments    []*FuncValueDesc `json:"arguments"`
	ReturnValues []*FuncValueDesc `json:"return_values"`
}

type InterfaceDesc struct {
	InterfaceName string           `json:"interface_name"`
	InMessages    []*OperationDesc `json:"in_messages"`
	OutMessages   []*OperationDesc `json:"out_messages"`
	GetMethods    []*GetMethodDesc `json:"get_methods"`
}
