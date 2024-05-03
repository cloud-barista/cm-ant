package jmetermanager

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
)

type jmxTemplateData struct {
	TestName          string
	Duration          string
	RampUpSteps       string
	RampUpTime        string
	VirtualUsers      string
	HttpRequests      string
	AgentHost         string
	AgentPort         string
	CpuResultPath     string
	MemoryResultPath  string
	SwapResultPath    string
	DiskResultPath    string
	NetworkResultPath string
	TcpResultPath     string
}

type jmxHttpTemplateData struct {
	Method   string                 `json:"method"`
	Protocol string                 `json:"protocol"`
	Hostname string                 `json:"hostname"`
	Port     string                 `json:"port"`
	Path     string                 `json:"path,omitempty"`
	Params   []jmxTemplateDataParam `json:"params,omitempty"`
	BodyData string                 `json:"bodyData,omitempty"`
}

type jmxTemplateDataParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var jmxHttpSamplerTemplate = map[string]string{
	"GET": `
	<HTTPSamplerProxy guiclass="HttpTestSampleGui" testclass="HTTPSamplerProxy" testname="{{.Method}} Request" enabled="true">
		<elementProp name="HTTPsampler.Arguments" elementType="Arguments" guiclass="HTTPArgumentsPanel" testclass="Arguments" testname="User Defined Variables" enabled="true">
			<collectionProp name="Arguments.arguments">
				{{ range .Params }}
				<elementProp name="{{ .Name }}" elementType="HTTPArgument">
					<boolProp name="HTTPArgument.always_encode">false</boolProp>
					<stringProp name="Argument.value">{{ .Value }}</stringProp>
					<stringProp name="Argument.metadata">=</stringProp>
					<boolProp name="HTTPArgument.use_equals">true</boolProp>
					<stringProp name="Argument.name">{{ .Name }}</stringProp>
				</elementProp>
				{{ end }}
			</collectionProp>
		</elementProp>
		<stringProp name="HTTPSampler.domain">{{.Hostname}}</stringProp>
		<stringProp name="HTTPSampler.port">{{.Port}}</stringProp>
		<stringProp name="HTTPSampler.protocol">{{.Protocol}}</stringProp>
		<stringProp name="HTTPSampler.path">{{.Path}}</stringProp>
		<stringProp name="HTTPSampler.method">{{.Method}}</stringProp>
		<stringProp name="HTTPSampler.contentEncoding">UTF-8</stringProp>
		<boolProp name="HTTPSampler.follow_redirects">true</boolProp>
		<boolProp name="HTTPSampler.auto_redirects">false</boolProp>
		<boolProp name="HTTPSampler.use_keepalive">true</boolProp>
		<boolProp name="HTTPSampler.DO_MULTIPART_POST">false</boolProp>
		<stringProp name="HTTPSampler.embedded_url_re"></stringProp>
		<stringProp name="HTTPSampler.connect_timeout">60000</stringProp>
		<stringProp name="HTTPSampler.response_timeout">60000</stringProp>
	</HTTPSamplerProxy>
	<hashTree/>
	`,
	"POST": `
	<HTTPSamplerProxy guiclass="HttpTestSampleGui" testclass="HTTPSamplerProxy" testname="{{.Method}} Request" enabled="true">
		<boolProp name="HTTPSampler.postBodyRaw">true</boolProp>
		<elementProp name="HTTPsampler.Arguments" elementType="Arguments">
			<collectionProp name="Arguments.arguments">
				<elementProp name="" elementType="HTTPArgument">
					<boolProp name="HTTPArgument.always_encode">false</boolProp>
					<stringProp name="Argument.value">{{.BodyData}}</stringProp>
					<stringProp name="Argument.metadata">=</stringProp>
				</elementProp>
			</collectionProp>
		</elementProp>
		<stringProp name="HTTPSampler.domain">{{.Hostname}}</stringProp>
		<stringProp name="HTTPSampler.port">{{.Port}}</stringProp>
		<stringProp name="HTTPSampler.protocol">{{.Protocol}}</stringProp>
		<stringProp name="HTTPSampler.path">{{.Path}}</stringProp>
		<stringProp name="HTTPSampler.method">{{.Method}}</stringProp>
		<stringProp name="HTTPSampler.contentEncoding">UTF-8</stringProp>
		<boolProp name="HTTPSampler.follow_redirects">true</boolProp>
		<boolProp name="HTTPSampler.auto_redirects">false</boolProp>
		<boolProp name="HTTPSampler.use_keepalive">true</boolProp>
		<boolProp name="HTTPSampler.DO_MULTIPART_POST">false</boolProp>
		<stringProp name="HTTPSampler.embedded_url_re"></stringProp>
		<stringProp name="HTTPSampler.connect_timeout">60000</stringProp>
		<stringProp name="HTTPSampler.response_timeout">60000</stringProp>
	</HTTPSamplerProxy>
	`,
}

func tearDown(jmeterPath, loadTestKey string) error {
	return os.Remove(fmt.Sprintf("%s/test_plan/%s.jmx", jmeterPath, loadTestKey))
}

func createTestPlanJmx(createdPath string, loadTestReq *api.LoadExecutionConfigReq) error {
	jmeterConf := configuration.Get().Load.JMeter
	resultPath := fmt.Sprintf("%s/result", jmeterConf.WorkDir)

	httpRequests, err := httpReqParseToJmx(loadTestReq.HttpReqs)
	if err != nil {
		return err
	}

	jmxTemplateData := jmxTemplateData{
		TestName:          loadTestReq.TestName,
		Duration:          loadTestReq.Duration,
		RampUpSteps:       loadTestReq.RampUpSteps,
		RampUpTime:        loadTestReq.RampUpTime,
		VirtualUsers:      loadTestReq.VirtualUsers,
		HttpRequests:      httpRequests,
		AgentHost:         loadTestReq.HttpReqs[0].Hostname,
		AgentPort:         "5555",
		CpuResultPath:     fmt.Sprintf("%s/%s_cpu_result.csv", resultPath, loadTestReq.LoadTestKey),
		MemoryResultPath:  fmt.Sprintf("%s/%s_memory_result.csv", resultPath, loadTestReq.LoadTestKey),
		SwapResultPath:    fmt.Sprintf("%s/%s_swap_result.csv", resultPath, loadTestReq.LoadTestKey),
		DiskResultPath:    fmt.Sprintf("%s/%s_disk_result.csv", resultPath, loadTestReq.LoadTestKey),
		NetworkResultPath: fmt.Sprintf("%s/%s_network_result.csv", resultPath, loadTestReq.LoadTestKey),
		TcpResultPath:     fmt.Sprintf("%s/%s_tcp_result.csv", resultPath, loadTestReq.LoadTestKey),
	}

	tmpl, err := template.ParseFiles(configuration.JoinRootPathWith("/test_plan/default_perfmon.jmx"))
	if err != nil {
		return err
	}

	outputFile, err := os.Create(fmt.Sprintf("%s/test_plan/%s.jmx", createdPath, loadTestReq.LoadTestKey))
	if err != nil {
		return err
	}
	defer outputFile.Close()

	err = tmpl.Execute(outputFile, jmxTemplateData)
	if err != nil {
		return err
	}

	return nil

}

func httpReqParseToJmx(httpReqs []api.LoadExecutionHttpReq) (string, error) {
	var builder strings.Builder
	for i, req := range httpReqs {
		jmxTemplate, ok := jmxHttpSamplerTemplate[req.Method]

		if ok {
			tmpl, err := template.New(fmt.Sprintf("jmxTemplate-%d-%s", i, req.Method)).Parse(jmxTemplate)
			if err != nil {
				return "", err
			}

			parsedUrl, err := url.Parse(req.Path)
			if err != nil {
				return "", err
			}

			jmxHttpTemplateData := jmxHttpTemplateData{
				Method:   req.Method,
				Protocol: req.Protocol,
				Hostname: req.Hostname,
				Port:     req.Port,
				Path:     parsedUrl.Path,
			}

			if req.Method == "GET" {
				params := make([]jmxTemplateDataParam, 0)
				parsedParams := parsedUrl.Query()
				for key, values := range parsedParams {
					param := jmxTemplateDataParam{
						Name:  key,
						Value: values[0],
					}

					params = append(params, param)
				}
				jmxHttpTemplateData.Params = params

			} else if req.Method == "POST" {
				jmxHttpTemplateData.BodyData = req.BodyData
			}

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, jmxHttpTemplateData)
			if err != nil {
				log.Fatalf("Error executing template: %s", err)
			}
			builder.WriteString(buf.String())
		}

		builder.WriteString("\n")
	}
	result := builder.String()
	return result, nil
}
