package spider

type PriceInfoReq struct {
	ConnectionName string      `json:"ConnectionName"`
	FilterList     []FilterReq `json:"FilterList"`
}

type FilterReq struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}
