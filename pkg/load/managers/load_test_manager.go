package managers

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/outbound"
	"log"
	"os"
	"strings"

	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/load/domain"
	"github.com/cloud-barista/cm-ant/pkg/utils"
)

type LoadTestManager struct {
}

func NewLoadTestManager() LoadTestManager {
	return LoadTestManager{}
}

func (l *LoadTestManager) Install(installReq domain.LoadEnvReq) error {
	installScriptPath := configuration.JoinRootPathWith("/script/install-jmeter.sh")

	if installReq.Type == domain.REMOTE {

		data, err := os.ReadFile(installScriptPath)
		if err != nil {
			log.Println("file doesn't exist on correct path")
			return err
		}

		multiLineCommand := strings.Replace(string(data), "#!/bin/bash", "", 1)

		commandReq := outbound.SendCommandReq{
			Command:  []string{multiLineCommand},
			UserName: installReq.Username,
		}

		stdout, err := outbound.SendCommandTo(configuration.TumblebugHostWithPort(), installReq.NsId, installReq.McisId, commandReq)

		if err != nil {
			log.Println(stdout)
			return err
		}

		log.Println(stdout)

	} else if installReq.Type == domain.LOCAL {

		err := utils.Script(installScriptPath, []string{
			fmt.Sprintf("JMETER_WORK_DIR=%s", configuration.Get().Load.JMeter.WorkDir),
			fmt.Sprintf("JMETER_VERSION=%s", configuration.Get().Load.JMeter.Version),
		})
		if err != nil {
			return fmt.Errorf("error while installing jmeter; %s", err)
		}

	}

	return nil
}

func (l *LoadTestManager) Run(property domain.LoadTestPropertyReq) (string, error) {
	var testId string
	testFolderSetupScript := configuration.JoinRootPathWith("/script/pre-execute-jmeter.sh")
	testPlanName := "test_plan_1.jmx"
	jmeterPath := configuration.Get().Load.JMeter.WorkDir
	jmeterVersion := configuration.Get().Load.JMeter.Version

	// TODO code cloud test using tumblebug
	if property.LoadEnvReq.Type == domain.REMOTE {
		tumblebugUrl := configuration.TumblebugHostWithPort()

		// 1. Installation check
		err := l.Install(property.LoadEnvReq)

		if err != nil {
			return "", err
		}

		// 2. pre-requirement check
		data, err := os.ReadFile(testFolderSetupScript)
		if err != nil {
			log.Println("file doesn't exist on correct path")
			return "", err
		}

		multiLineCommand := strings.Replace(string(data), "#!/bin/bash", "", 1)
		multiLineCommand = strings.Replace(multiLineCommand, "$1", "testPlanName", 1)

		commandReq := outbound.SendCommandReq{
			Command:  []string{multiLineCommand},
			UserName: property.LoadEnvReq.Username,
		}

		stdout, err := outbound.SendCommandTo(tumblebugUrl, property.LoadEnvReq.NsId, property.LoadEnvReq.McisId, commandReq)

		if err != nil {
			log.Printf("error occured; %s\n", err)
			log.Println(stdout)
			return "", err
		}

		log.Println(stdout)

		// 3. execute jmeter test
		jmeterTestCommand := executionCmdGen(property, testPlanName, fmt.Sprintf("%s_result.csv", property.PropertiesId))

		commandReq = outbound.SendCommandReq{
			Command:  []string{jmeterTestCommand},
			UserName: property.LoadEnvReq.Username,
		}

		stdout, err = outbound.SendCommandTo(tumblebugUrl, property.LoadEnvReq.NsId, property.LoadEnvReq.McisId, commandReq)

		if err != nil {
			log.Printf("error occured; %s\n", err)
			log.Println(stdout)
			return "", err
		}

		log.Println(stdout)

	} else if property.LoadEnvReq.Type == domain.LOCAL {

		log.Printf("[%s] Do load test on local", property.PropertiesId)

		exist := utils.ExistCheck(jmeterPath)

		if !exist {
			loadInstallReq := domain.LoadEnvReq{
				Type: domain.LOCAL,
			}

			err := l.Install(loadInstallReq)

			if err != nil {
				log.Printf("error while execute [Run()]; %s\n", err)
				return "", err
			}
		}

		err := utils.Script(testFolderSetupScript, []string{
			fmt.Sprintf("TEST_PLAN_NAME=%s", testPlanName),
			fmt.Sprintf("JMETER_WORK_DIR=%s", jmeterPath),
			fmt.Sprintf("JMETER_VERSION=%s", jmeterVersion),
		})

		if err != nil {
			log.Println(err)
			return "", err
		}

		jmeterTestCommand := executionCmdGen(property, testPlanName, fmt.Sprintf("%s_result.csv", property.PropertiesId))
		err = utils.InlineCmd(jmeterTestCommand)

		if err != nil {
			log.Println(err)
			return "", err
		}

		// 3. save test configuration
		testId = property.PropertiesId
	}

	return testId, nil
}

func executionCmdGen(p domain.LoadTestPropertyReq, testPlanName, resultFileName string) string {
	jmeterConf := configuration.Get().Load.JMeter

	var builder strings.Builder
	testPath := fmt.Sprintf("%s/test_plan/%s", jmeterConf.WorkDir, testPlanName)
	resultPath := fmt.Sprintf("%s/result/%s", jmeterConf.WorkDir, resultFileName)

	builder.WriteString(fmt.Sprintf("%s/apache-jmeter-%s/bin/jmeter.sh", jmeterConf.WorkDir, jmeterConf.Version))
	builder.WriteString(" -n -f")
	builder.WriteString(fmt.Sprintf(" -Jthreads=%s", p.Threads))
	builder.WriteString(fmt.Sprintf(" -JrampTime=%s", p.RampTime))
	builder.WriteString(fmt.Sprintf(" -JloopCount=%s", p.LoopCount))
	builder.WriteString(fmt.Sprintf(" -Jprotocol=%s", p.HttpReqs.Protocol))
	builder.WriteString(fmt.Sprintf(" -Jhostname=%s", p.HttpReqs.Hostname))
	builder.WriteString(fmt.Sprintf(" -Jport=%s", p.HttpReqs.Port))
	builder.WriteString(fmt.Sprintf(" -Jpath=%s", p.HttpReqs.Path))
	builder.WriteString(fmt.Sprintf(" -JbodyData=%s", p.HttpReqs.BodyData))
	builder.WriteString(fmt.Sprintf(" -JbodyData=%s", p.LoopCount))
	builder.WriteString(fmt.Sprintf(" -t=%s", testPath))
	builder.WriteString(fmt.Sprintf(" -l=%s", resultPath))

	return builder.String()
}
