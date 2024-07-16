package load

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"text/template"

	"github.com/cloud-barista/cm-ant/pkg/utils"
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

func parseTestPlanStructToString(w io.Writer, param RunLoadTestParam, loadGeneratorInstallInfo *LoadGeneratorInstallInfo) error {
	resultPath := fmt.Sprintf("%s/result", loadGeneratorInstallInfo.InstallPath)
	httpRequests, err := httpReqParseToJmx(param.Hostname, param.Port, param.HttpReqs)
	if err != nil {
		return err
	}

	agentHost := param.AgentHostname

	var tmpl *template.Template

	jmxTemplateData := jmxTemplateData{
		TestName:     param.TestName,
		Duration:     param.Duration,
		RampUpSteps:  param.RampUpSteps,
		RampUpTime:   param.RampUpTime,
		VirtualUsers: param.VirtualUsers,
		HttpRequests: httpRequests,
	}

	if agentHost != "" {
		jmxTemplateData.AgentHost = agentHost
		jmxTemplateData.AgentPort = "5555"
		jmxTemplateData.CpuResultPath = fmt.Sprintf("%s/%s_cpu_result.csv", resultPath, param.LoadTestKey)
		jmxTemplateData.MemoryResultPath = fmt.Sprintf("%s/%s_memory_result.csv", resultPath, param.LoadTestKey)
		jmxTemplateData.DiskResultPath = fmt.Sprintf("%s/%s_disk_result.csv", resultPath, param.LoadTestKey)
		jmxTemplateData.NetworkResultPath = fmt.Sprintf("%s/%s_network_result.csv", resultPath, param.LoadTestKey)

		tmpl, err = template.ParseFiles(utils.JoinRootPathWith("/test_plan/default_perfmon.jmx"))
	} else {
		tmpl, err = template.ParseFiles(utils.JoinRootPathWith("/test_plan/default.jmx"))
	}

	if err != nil {
		return err
	}

	err = tmpl.Execute(w, jmxTemplateData)
	if err != nil {
		return err
	}

	return nil
}

func httpReqParseToJmx(hostname, port string, httpReqs []RunLoadTestHttpParam) (string, error) {
	var builder strings.Builder
	for i, req := range httpReqs {
		method := strings.ToUpper(req.Method)
		jmxTemplate, ok := jmxHttpSamplerTemplate[method]

		if ok {
			tmpl, err := template.New(fmt.Sprintf("jmxTemplate-%d-%s", i, method)).Parse(jmxTemplate)
			if err != nil {
				return "", err
			}

			parsedUrl, err := url.Parse(req.Path)
			if err != nil {
				return "", err
			}

			h := req.Hostname
			if h == "" {
				h = hostname
			}

			p := req.Port
			if p == "" {
				p = port
			}

			jmxHttpTemplateData := jmxHttpTemplateData{
				Method:   method,
				Protocol: req.Protocol,
				Hostname: h,
				Port:     p,
				Path:     parsedUrl.Path,
			}

			if strings.ToUpper(method) == "GET" {
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
			} else if method == "POST" {
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
