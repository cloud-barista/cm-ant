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
	TestName     string `json:"testName"`
	Duration     string `json:"duration"`
	RampUpSteps  string `json:"rampUpSteps"`
	RampUpTime   string `json:"rampUpTime"`
	VirtualUsers string `json:"virtualUsers"`
	HttpRequests string `json:"httpRequests"`
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

const defaultJmx = `
<?xml version="1.0" encoding="UTF-8"?>
<jmeterTestPlan version="1.2" properties="5.0" jmeter="5.3">
  <hashTree>
    <TestPlan guiclass="TestPlanGui" testclass="TestPlan" testname="{{.TestName}}" enabled="true">
      <stringProp name="TestPlan.comments"></stringProp>
      <boolProp name="TestPlan.functional_mode">false</boolProp>
      <boolProp name="TestPlan.tearDown_on_shutdown">true</boolProp>
      <boolProp name="TestPlan.serialize_threadgroups">false</boolProp>
      <elementProp name="TestPlan.user_defined_variables" elementType="Arguments" guiclass="ArgumentsPanel" testclass="Arguments" testname="User Defined Variables" enabled="true">
        <collectionProp name="Arguments.arguments"/>
      </elementProp>
      <stringProp name="TestPlan.user_define_classpath"></stringProp>
    </TestPlan>
    <hashTree>
      <com.blazemeter.jmeter.threads.concurrency.ConcurrencyThreadGroup guiclass="com.blazemeter.jmeter.threads.concurrency.ConcurrencyThreadGroupGui" testclass="com.blazemeter.jmeter.threads.concurrency.ConcurrencyThreadGroup" testname="{{.TestName}} Thread Group" enabled="true">
        <elementProp name="ThreadGroup.main_controller" elementType="com.blazemeter.jmeter.control.VirtualUserController"/>
        <stringProp name="Hold">{{.Duration}}</stringProp>
        <stringProp name="Steps">{{.RampUpSteps}}</stringProp>
        <stringProp name="RampUp">{{.RampUpTime}}</stringProp>
        <stringProp name="TargetLevel">{{.VirtualUsers}}</stringProp>
        <stringProp name="Iterations">0</stringProp>
        <stringProp name="Unit">S</stringProp>
        <stringProp name="ThreadGroup.on_sample_error">continue</stringProp>
      </com.blazemeter.jmeter.threads.concurrency.ConcurrencyThreadGroup>
      <hashTree>
        {{.HttpRequests}}
      </hashTree>
      <hashTree/>
      <HeaderManager guiclass="HeaderPanel" testclass="HeaderManager" testname="HTTP Header Manager" enabled="true">
        <collectionProp name="HeaderManager.headers">
          <elementProp name="" elementType="Header">
            <stringProp name="Header.name">test-header</stringProp>
            <stringProp name="Header.value">test-header-value</stringProp>
          </elementProp>
        </collectionProp>
      </HeaderManager>
      <hashTree/>
      <CookieManager guiclass="CookiePanel" testclass="CookieManager" testname="HTTP Cookie Manager" enabled="true">
        <collectionProp name="CookieManager.cookies">
          <elementProp name="test-cookie" elementType="Cookie" testname="test-cookie">
            <stringProp name="Cookie.value">test-cookie-value</stringProp>
            <stringProp name="Cookie.domain">test.domain.com</stringProp>
            <stringProp name="Cookie.path"></stringProp>
            <boolProp name="Cookie.secure">false</boolProp>
            <longProp name="Cookie.expires">0</longProp>
            <boolProp name="Cookie.path_specified">true</boolProp>
            <boolProp name="Cookie.domain_specified">true</boolProp>
          </elementProp>
        </collectionProp>
        <boolProp name="CookieManager.clearEachIteration">false</boolProp>
        <boolProp name="CookieManager.controlledByThreadGroup">false</boolProp>
      </CookieManager>
      <hashTree/>
      <CacheManager guiclass="CacheManagerGui" testclass="CacheManager" testname="HTTP Cache Manager" enabled="true">
        <boolProp name="clearEachIteration">false</boolProp>
        <boolProp name="useExpires">true</boolProp>
        <boolProp name="CacheManager.controlledByThread">false</boolProp>
      </CacheManager>
      <hashTree/>
    </hashTree>
  </hashTree>
</jmeterTestPlan>
`

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

	httpRequests, err := httpReqParseToJmx(loadTestReq.HttpReqs)

	if err != nil {
		return err
	}

	jmxTemplateData := jmxTemplateData{
		TestName:     loadTestReq.TestName,
		Duration:     loadTestReq.Duration,
		RampUpSteps:  loadTestReq.RampUpSteps,
		RampUpTime:   loadTestReq.RampUpTime,
		VirtualUsers: loadTestReq.VirtualUsers,
		HttpRequests: httpRequests,
	}

	tmpl, err := template.ParseFiles(configuration.JoinRootPathWith("/test_plan/default.jmx"))
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
