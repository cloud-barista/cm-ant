package jmetermanager

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/pkg/config"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/outbound/tumblebug"
	"github.com/cloud-barista/cm-ant/pkg/utils"
)

type JMeterLoadTestManager struct {
}

func (j *JMeterLoadTestManager) Install(loadEnvReq *api.LoadEnvReq) error {
	installScriptPath := utils.JoinRootPathWith("/script/install-jmeter.sh")

	if loadEnvReq.InstallLocation == constant.Remote {
		installationCommand, err := utils.ReadToString(installScriptPath)
		if err != nil {
			log.Println("file doesn't exist on correct path")
			return err
		}

		commandReq := tumblebug.SendCommandReq{
			Command:  []string{installationCommand},
			UserName: loadEnvReq.Username,
		}

		stdout, err := tumblebug.CommandToVm(loadEnvReq.NsId, loadEnvReq.McisId, loadEnvReq.VmId, commandReq)

		if err != nil {
			log.Println("error! ", stdout)
			return err
		}

		log.Println("command result", stdout)

		err = utils.AddToKnownHost(loadEnvReq.PemKeyPath, loadEnvReq.PublicIp, loadEnvReq.Username)
		if err != nil {
			return err
		}

	} else if loadEnvReq.InstallLocation == constant.Local {

		err := utils.Script(installScriptPath, []string{
			fmt.Sprintf("JMETER_WORK_DIR=%s", config.AppConfig.Load.JMeter.Dir),
			fmt.Sprintf("JMETER_VERSION=%s", config.AppConfig.Load.JMeter.Version),
		})
		if err != nil {
			return fmt.Errorf("error while installing jmeter; %s", err)
		}
	}

	return nil
}

func (j *JMeterLoadTestManager) Uninstall(loadEnvReq *api.LoadEnvReq) error {
	uninstallScriptPath := utils.JoinRootPathWith("/script/uninstall-jmeter.sh")

	if loadEnvReq.InstallLocation == constant.Remote {
		uninstallCommand, err := utils.ReadToString(uninstallScriptPath)
		if err != nil {
			log.Println("file doesn't exist on correct path")
			return err
		}

		commandReq := tumblebug.SendCommandReq{
			Command:  []string{uninstallCommand},
			UserName: loadEnvReq.Username,
		}

		stdout, err := tumblebug.CommandToVm(loadEnvReq.NsId, loadEnvReq.McisId, loadEnvReq.VmId, commandReq)

		if err != nil {
			log.Println(stdout)
			return err
		}

		log.Println(stdout)

	} else if loadEnvReq.InstallLocation == constant.Local {

		err := utils.Script(uninstallScriptPath, []string{
			fmt.Sprintf("JMETER_WORK_DIR=%s", config.AppConfig.Load.JMeter.Dir),
			fmt.Sprintf("JMETER_VERSION=%s", config.AppConfig.Load.JMeter.Version),
		})
		if err != nil {
			return fmt.Errorf("error while installing jmeter; %s", err)
		}
	}

	return nil
}

func (j *JMeterLoadTestManager) Stop(loadTestReq api.LoadExecutionConfigReq) error {
	killCmd := killCmdGen(loadTestReq.LoadTestKey)

	// TODO code cloud test using tumblebug
	loadEnv := loadTestReq.LoadEnvReq
	if loadEnv.InstallLocation == constant.Remote {

		commandReq := tumblebug.SendCommandReq{
			Command:  []string{killCmd},
			UserName: loadEnv.Username,
		}

		stdout, err := tumblebug.CommandToVm(loadEnv.NsId, loadEnv.McisId, loadEnv.VmId, commandReq)

		if err != nil {
			log.Println(stdout)
			return err
		}

		log.Println(stdout)

	} else if loadEnv.InstallLocation == constant.Local {
		log.Printf("[%s] stop load test on local", loadTestReq.LoadTestKey)

		err := utils.InlineCmd(killCmd)

		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (j *JMeterLoadTestManager) Run(loadTestReq *api.LoadExecutionConfigReq) error {
	checkRequirementPath := utils.JoinRootPathWith("/script/pre-execute-jmeter.sh")
	testPlanName := fmt.Sprintf("%s.jmx", loadTestReq.LoadTestKey)
	jmeterPath := config.AppConfig.Load.JMeter.Dir
	jmeterVersion := config.AppConfig.Load.JMeter.Version
	loadEnv := loadTestReq.LoadEnvReq
	resultFileName := fmt.Sprintf("%s_result.csv", loadTestReq.LoadTestKey)

	if loadEnv.InstallLocation == constant.Remote {
		preRequirementCmd, err := utils.ReadToString(checkRequirementPath)
		if err != nil {
			log.Println("file doesn't exist on correct path")
			return err
		}
		preRequirementCmd = strings.Replace(preRequirementCmd, "${TEST_PLAN_NAME}", testPlanName, 1)

		testPlan, err := TestPlanJmx(loadTestReq)
		if err != nil {
			return err
		}
		encoded := base64.StdEncoding.EncodeToString([]byte(testPlan))

		createFileCmd := fmt.Sprintf("echo \"%s\" | base64 -d >> %s/test_plan/%s", encoded, jmeterPath, testPlanName)
		commandReq := tumblebug.SendCommandReq{
			Command:  []string{createFileCmd},
			UserName: loadTestReq.LoadEnvReq.Username,
		}
		stdout, err := tumblebug.CommandToVm(loadTestReq.LoadEnvReq.NsId, loadTestReq.LoadEnvReq.McisId, loadTestReq.LoadEnvReq.VmId, commandReq)

		if err != nil {
			log.Printf("error occured; %s\n", err)
			log.Println(stdout)
			return err
		}

		jmeterTestCommand := executionCmd(testPlanName, resultFileName)

		commandReq = tumblebug.SendCommandReq{
			Command:  []string{jmeterTestCommand},
			UserName: loadTestReq.LoadEnvReq.Username,
		}

		stdout, err = tumblebug.CommandToVm(loadTestReq.LoadEnvReq.NsId, loadTestReq.LoadEnvReq.McisId, loadTestReq.LoadEnvReq.VmId, commandReq)

		if err != nil {
			log.Printf("error occured; %s\n", err)
			log.Println(stdout)
			return err
		}

		if strings.Contains(stdout, "exited with status 1") {
			return errors.New("error with load test")
		}

	} else if loadEnv.InstallLocation == constant.Local {

		log.Printf("[%s] do load test on local", loadTestReq.LoadTestKey)

		// 1. jmeter installation check
		exist := utils.ExistCheck(jmeterPath)

		if !exist {
			loadInstallReq := api.LoadEnvReq{
				InstallLocation: constant.Local,
			}

			err := j.Install(&loadInstallReq)

			if err != nil {
				log.Printf("error while execute [Run()]; %s\n", err)
				return err
			}
		}

		// 2. jmx file create
		err := createTestPlanJmx(jmeterPath, loadTestReq)
		if err != nil {
			return err
		}

		defer tearDown(jmeterPath, loadTestReq.LoadTestKey)

		// 3. check pre requirement
		err = utils.Script(checkRequirementPath, []string{
			fmt.Sprintf("TEST_PLAN_NAME=%s", testPlanName),
			fmt.Sprintf("JMETER_WORK_DIR=%s", jmeterPath),
			fmt.Sprintf("JMETER_VERSION=%s", jmeterVersion),
		})

		if err != nil {
			log.Println(err)
			return err
		}

		// 4. execution jmeter test
		jmeterTestCommand := executionCmd(testPlanName, resultFileName)
		err = utils.InlineCmd(jmeterTestCommand)

		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func executionCmd(testPlanName, resultFileName string) string {
	jmeterConf := config.AppConfig.Load.JMeter

	var builder strings.Builder
	testPath := fmt.Sprintf("%s/test_plan/%s", jmeterConf.Dir, testPlanName)
	resultPath := fmt.Sprintf("%s/result/%s", jmeterConf.Dir, resultFileName)

	builder.WriteString(fmt.Sprintf("%s/apache-jmeter-%s/bin/jmeter.sh", jmeterConf.Dir, jmeterConf.Version))
	builder.WriteString(" -n -f")
	builder.WriteString(fmt.Sprintf(" -t=%s", testPath))
	builder.WriteString(fmt.Sprintf(" -l=%s", resultPath))

	builder.WriteString(fmt.Sprintf(" && sudo rm %s", testPath))
	return builder.String()
}

func killCmdGen(loadTestKey string) string {
	grepRegex := fmt.Sprintf("'\\/bin\\/ApacheJMeter\\.jar.*%s'", loadTestKey)

	return fmt.Sprintf("kill -15 $(ps -ef | grep -E %s | awk '{print $2}')", grepRegex)
}
