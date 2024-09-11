package spider

type PriceInfoReq struct {
	ConnectionName string      `json:"ConnectionName"`
	FilterList     []FilterReq `json:"FilterList"`
}

type FilterReq struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type AnycallReq struct {
	ConnectionName string  `json:"ConnectionName"`
	ReqInfo        ReqInfo `json:"ReqInfo"`
}
type ReqInfo struct {
	FID           string     `json:"FID"`
	IKeyValueList []KeyValue `json:"IKeyValueList"`
}

type KeyValue struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}
