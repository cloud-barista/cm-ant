package cost

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/spider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadSpiderPriceFixture(t *testing.T, name string) spider.CloudPriceDataRes {
	t.Helper()

	b, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)

	var res spider.CloudPriceDataRes
	require.NoError(t, json.Unmarshal(b, &res))
	return res
}

func TestBuildEstimateCostInfos_GCPOnDemand(t *testing.T) {
	s := &SpiderPriceCollector{}
	res := loadSpiderPriceFixture(t, "spider_price_gcp_e2.json")
	param := RecommendSpecParam{ProviderName: "gcp", RegionName: "asia-northeast3", InstanceType: "e2-standard-2"}

	got := s.buildEstimateCostInfos(param, res)

	require.Len(t, got, 1)
	pi := got[0]
	assert.Equal(t, "0.08596572", pi.Price)
	assert.Equal(t, constant.PerHour, pi.Unit)
	assert.Equal(t, constant.USD, pi.Currency)
	assert.Equal(t, "asia-northeast3", pi.RegionName)
	assert.Equal(t, "e2-standard-2", pi.InstanceType)
	assert.Equal(t, "2", pi.VCpu)
	assert.Equal(t, "8", pi.Memory)
	assert.Equal(t, constant.GIB, pi.MemoryUnit)
	assert.Equal(t, "8192", pi.OriginalMemory)
	assert.InDelta(t, 0.08596572*24*30, pi.CalculatedMonthlyPrice, 1e-6)
}

func TestBuildEstimateCostInfos_AWSTermsExcludesReservation(t *testing.T) {
	s := &SpiderPriceCollector{}
	res := loadSpiderPriceFixture(t, "spider_price_aws_t3small.json")
	param := RecommendSpecParam{ProviderName: "aws", RegionName: "ap-northeast-2", InstanceType: "t3.small"}

	got := s.buildEstimateCostInfos(param, res)

	require.Len(t, got, 1)
	pi := got[0]
	price, err := strconv.ParseFloat(pi.Price, 64)
	require.NoError(t, err)
	assert.InDelta(t, 0.026, price, 1e-9)
	assert.Equal(t, constant.PerHour, pi.Unit)
	assert.Equal(t, "ap-northeast-2", pi.RegionName)
	assert.Equal(t, "2", pi.Memory)
	assert.Equal(t, constant.GIB, pi.MemoryUnit)
	assert.NotContains(t, pi.PriceDescription, "Reservation")
}

func TestBuildEstimateCostInfos_NonAWSIgnoresTermsStructure(t *testing.T) {
	s := &SpiderPriceCollector{}
	res := loadSpiderPriceFixture(t, "spider_price_gcp_e2.json")
	// A non-AWS CSPPriceInfo that happens to look like the AWS price list must not be read as AWS terms.
	res.PriceList[0].PriceInfo.CSPPriceInfo = map[string]interface{}{
		"terms": map[string]interface{}{
			"OnDemand": map[string]interface{}{
				"X": map[string]interface{}{
					"priceDimensions": map[string]interface{}{
						"Y": map[string]interface{}{
							"description":  "not a gcp price",
							"pricePerUnit": map[string]interface{}{"USD": "9.99"},
						},
					},
				},
			},
		},
	}
	param := RecommendSpecParam{ProviderName: "gcp", RegionName: "asia-northeast3", InstanceType: "e2-standard-2"}

	got := s.buildEstimateCostInfos(param, res)

	require.Len(t, got, 1)
	assert.Equal(t, "0.08596572", got[0].Price)
	assert.Equal(t, "OnDemand pricing for e2-standard-2", got[0].PriceDescription)
}

func TestBuildEstimateCostInfos_InstanceTypeMismatch(t *testing.T) {
	s := &SpiderPriceCollector{}
	res := loadSpiderPriceFixture(t, "spider_price_gcp_e2.json")
	param := RecommendSpecParam{ProviderName: "gcp", RegionName: "asia-northeast3", InstanceType: "e2-micro"}

	got := s.buildEstimateCostInfos(param, res)

	assert.Empty(t, got)
}

func TestResolveRegion(t *testing.T) {
	assert.Equal(t, "asia-northeast3", resolveRegion(nil, "asia-northeast3"))
	assert.Equal(t, "asia-northeast3", resolveRegion(&spider.VMSpecInfoRes{Region: ""}, "asia-northeast3"))
	assert.Equal(t, "ap-northeast-2", resolveRegion(&spider.VMSpecInfoRes{Region: "ap-northeast-2"}, "requested"))
}

func TestMibToGiB(t *testing.T) {
	s := &SpiderPriceCollector{}

	cases := []struct {
		in       string
		wantNum  string
		wantUnit constant.MemoryUnit
	}{
		{"8192", "8", constant.GIB},
		{"2048", "2", constant.GIB},
		{"1740.8", "1.7", constant.GIB},
		{"", "", ""},
		{"abc", "", ""},
	}

	for _, c := range cases {
		num, unit := s.mibToGiB(c.in)
		assert.Equal(t, c.wantNum, num, "input %q", c.in)
		assert.Equal(t, c.wantUnit, unit, "input %q", c.in)
	}
}
