package jmetermanager

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/constant"
	"github.com/cloud-barista/cm-ant/pkg/outbound/tumblebug"
	"github.com/melbahja/goph"

	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/utils"
)

type JMeterLoadTestManager struct {
}

func (j *JMeterLoadTestManager) Install(loadEnvReq *api.LoadEnvReq) error {
	installScriptPath := configuration.JoinRootPathWith("/script/install-jmeter.sh")

	if loadEnvReq.InstallLocation == constant.Remote {
		installationCommand, err := readAndParseScript(installScriptPath)
		if err != nil {
			log.Println("file doesn't exist on correct path")
			return err
		}

		switch loadEnvReq.RemoteConnectionType {
		case constant.BuiltIn:

			commandReq := tumblebug.SendCommandReq{
				Command:  []string{installationCommand},
				UserName: loadEnvReq.Username,
			}

			stdout, err := tumblebug.CommandToMcis(loadEnvReq.NsId, loadEnvReq.McisId, commandReq)

			if err != nil {
				log.Println(stdout)
				return err
			}

			log.Println(stdout)
		case constant.PrivateKey, constant.Password:
			var auth goph.Auth
			var err error

			if loadEnvReq.RemoteConnectionType == constant.PrivateKey {
				auth, err = goph.Key(loadEnvReq.Cert, "")
				if err != nil {
					return err
				}
			} else if loadEnvReq.RemoteConnectionType == constant.Password {
				auth = goph.Password(loadEnvReq.Cert)
				if err != nil {
					return err
				}
			}

			client, err := goph.New(loadEnvReq.Username, loadEnvReq.PublicIp, auth)
			if err != nil {
				return err
			}

			defer client.Close()

			out, err := client.RunContext(context.Background(), installationCommand)

			if err != nil {
				log.Println(string(out))
				return err
			}

			log.Println(string(out))
		}

	} else if loadEnvReq.InstallLocation == constant.Local {

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

func (j *JMeterLoadTestManager) Uninstall(loadEnvReq *api.LoadEnvReq) error {
	uninstallScriptPath := configuration.JoinRootPathWith("/script/uninstall-jmeter.sh")

	if loadEnvReq.InstallLocation == constant.Remote {
		uninstallCommand, err := readAndParseScript(uninstallScriptPath)
		if err != nil {
			log.Println("file doesn't exist on correct path")
			return err
		}

		switch loadEnvReq.RemoteConnectionType {
		case constant.BuiltIn:
			commandReq := tumblebug.SendCommandReq{
				Command:  []string{uninstallCommand},
				UserName: loadEnvReq.Username,
			}

			stdout, err := tumblebug.CommandToMcis(loadEnvReq.NsId, loadEnvReq.McisId, commandReq)

			if err != nil {
				log.Println(stdout)
				return err
			}

			log.Println(stdout)
		case constant.PrivateKey, constant.Password:
			var auth goph.Auth
			var err error

			if loadEnvReq.RemoteConnectionType == constant.PrivateKey {
				auth, err = goph.Key(loadEnvReq.Cert, "")
				if err != nil {
					return err
				}
			} else if loadEnvReq.RemoteConnectionType == constant.Password {
				auth = goph.Password(loadEnvReq.Cert)
			}

			client, err := goph.New(loadEnvReq.Username, loadEnvReq.PublicIp, auth)
			if err != nil {
				return err
			}

			defer client.Close()

			out, err := client.RunContext(context.Background(), uninstallCommand)

			if err != nil {
				log.Println(string(out))
				return err
			}

			log.Println(string(out))
		}

	} else if loadEnvReq.InstallLocation == constant.Local {

		err := utils.Script(uninstallScriptPath, []string{
			fmt.Sprintf("JMETER_WORK_DIR=%s", configuration.Get().Load.JMeter.WorkDir),
			fmt.Sprintf("JMETER_VERSION=%s", configuration.Get().Load.JMeter.Version),
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

		switch loadEnv.RemoteConnectionType {
		case constant.BuiltIn:
			commandReq := tumblebug.SendCommandReq{
				Command:  []string{killCmd},
				UserName: loadEnv.Username,
			}

			stdout, err := tumblebug.CommandToMcis(loadEnv.NsId, loadEnv.McisId, commandReq)

			if err != nil {
				log.Println(stdout)
				return err
			}

			log.Println(stdout)
		case constant.PrivateKey, constant.Password:
			var auth goph.Auth
			var err error

			if loadEnv.RemoteConnectionType == constant.PrivateKey {
				auth, err = goph.Key(loadEnv.Cert, "")
				if err != nil {
					return err
				}
			} else if loadEnv.RemoteConnectionType == constant.Password {
				auth = goph.Password(loadEnv.Cert)
			}

			// 1. ssh client connection
			client, err := goph.New(loadEnv.Username, loadEnv.PublicIp, auth)
			if err != nil {
				return err
			}

			defer client.Close()

			out, err := client.RunContext(context.Background(), killCmd)

			if err != nil {
				log.Println("error while kill cmd", string(out), err)
				return err
			}

			log.Println(string(out))
		}

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
	checkRequirementPath := configuration.JoinRootPathWith("/script/pre-execute-jmeter.sh")
	testPlanName := fmt.Sprintf("%s.jmx", loadTestReq.LoadTestKey)
	jmeterPath := configuration.Get().Load.JMeter.WorkDir
	jmeterVersion := configuration.Get().Load.JMeter.Version
	loadEnv := loadTestReq.LoadEnvReq
	resultFileName := fmt.Sprintf("%s_result.csv", loadTestReq.LoadTestKey)

	// TODO code cloud test using tumblebug
	if loadEnv.InstallLocation == constant.Remote {
		preRequirementCmd, err := readAndParseScript(checkRequirementPath)
		if err != nil {
			log.Println("file doesn't exist on correct path")
			return err
		}
		preRequirementCmd = strings.Replace(preRequirementCmd, "${TEST_PLAN_NAME}", testPlanName, 1)

		switch loadEnv.RemoteConnectionType {
		case constant.BuiltIn:
			commandReq := tumblebug.SendCommandReq{
				Command:  []string{preRequirementCmd},
				UserName: loadTestReq.LoadEnvReq.Username,
			}

			// 1. check pre-requisition
			stdout, err := tumblebug.CommandToMcis(loadTestReq.LoadEnvReq.NsId, loadTestReq.LoadEnvReq.McisId, commandReq)

			if err != nil {
				log.Printf("error occured; %s\n", err)
				log.Println(stdout)
				return err
			}

			log.Println(stdout)

			// 2. execute jmeter test
			jmeterTestCommand := executionCmdGen(loadTestReq, testPlanName, resultFileName)

			commandReq = tumblebug.SendCommandReq{
				Command:  []string{jmeterTestCommand},
				UserName: loadTestReq.LoadEnvReq.Username,
			}

			stdout, err = tumblebug.CommandToMcis(loadTestReq.LoadEnvReq.NsId, loadTestReq.LoadEnvReq.McisId, commandReq)

			if err != nil {
				log.Printf("error occured; %s\n", err)
				log.Println(stdout)
				return err
			}

			log.Println(stdout)
		case constant.PrivateKey, constant.Password:
			var auth goph.Auth
			var err error

			if loadEnv.RemoteConnectionType == constant.PrivateKey {
				auth, err = goph.Key(loadEnv.Cert, "")
				if err != nil {
					return err
				}
			} else if loadEnv.RemoteConnectionType == constant.Password {
				auth = goph.Password(loadEnv.Cert)
			}

			// 1. ssh sshClient connection
			sshClient, err := goph.New(loadEnv.Username, loadEnv.PublicIp, auth)
			if err != nil {
				return err
			}

			defer sshClient.Close()

			// 2. upload test file from ant to remote
			err = createTestPlanJmx(configuration.RootPath(), loadTestReq)
			if err != nil {
				return err
			}

			defer func() {
				err := tearDown(configuration.RootPath(), loadTestReq.LoadTestKey)
				if err == nil {
					out, err := sshClient.RunContext(context.Background(), fmt.Sprintf("rm %s/test_plan/%s", jmeterPath, testPlanName))

					if err != nil {
						log.Println(string(out))
					}
				}
			}()

			err = sshClient.Upload(fmt.Sprintf("%s/test_plan/%s", configuration.RootPath(), testPlanName), fmt.Sprintf("%s/test_plan/%s", jmeterPath, testPlanName))

			if err != nil {
				return err
			}

			// 3. check pre-requisition
			out, err := sshClient.RunContext(context.Background(), preRequirementCmd)

			if err != nil {
				log.Println(string(out))
				return err
			}

			log.Println(string(out))

			// 4. execute jmeter test
			jmeterTestCommand := executionCmd(testPlanName, resultFileName)
			out, err = sshClient.RunContext(context.Background(), jmeterTestCommand)

			if err != nil {
				log.Println(string(out))
				return err
			}

			log.Println(string(out))
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
	jmeterConf := configuration.Get().Load.JMeter

	var builder strings.Builder
	testPath := fmt.Sprintf("%s/test_plan/%s", jmeterConf.WorkDir, testPlanName)
	resultPath := fmt.Sprintf("%s/result/%s", jmeterConf.WorkDir, resultFileName)

	builder.WriteString(fmt.Sprintf("%s/apache-jmeter-%s/bin/jmeter.sh", jmeterConf.WorkDir, jmeterConf.Version))
	builder.WriteString(" -n -f")
	builder.WriteString(fmt.Sprintf(" -t=%s", testPath))
	builder.WriteString(fmt.Sprintf(" -l=%s", resultPath))

	return builder.String()
}

func executionCmdGen(p *api.LoadExecutionConfigReq, testPlanName, resultFileName string) string {
	jmeterConf := configuration.Get().Load.JMeter

	var builder strings.Builder
	testPath := fmt.Sprintf("%s/test_plan/%s", jmeterConf.WorkDir, testPlanName)
	resultPath := fmt.Sprintf("%s/result/%s", jmeterConf.WorkDir, resultFileName)

	builder.WriteString(fmt.Sprintf("%s/apache-jmeter-%s/bin/jmeter.sh", jmeterConf.WorkDir, jmeterConf.Version))
	builder.WriteString(" -n -f")
	builder.WriteString(fmt.Sprintf(" -Jthreads=%s", p.VirtualUsers))
	builder.WriteString(fmt.Sprintf(" -JrampTime=%s", p.Duration))
	builder.WriteString(fmt.Sprintf(" -JloopCount=%s", p.RampUpTime))
	builder.WriteString(fmt.Sprintf(" -Jprotocol=%s", p.HttpReqs[0].Protocol))
	builder.WriteString(fmt.Sprintf(" -Jhostname=%s", p.HttpReqs[0].Hostname))
	builder.WriteString(fmt.Sprintf(" -Jport=%s", p.HttpReqs[0].Port))
	builder.WriteString(fmt.Sprintf(" -Jpath=%s", p.HttpReqs[0].Path))
	builder.WriteString(fmt.Sprintf(" -JbodyData=%s", p.HttpReqs[0].BodyData))
	builder.WriteString(fmt.Sprintf(" -JbodyData=%s", p.RampUpTime))
	builder.WriteString(fmt.Sprintf(" -JpropertiesId=%s", p.LoadTestKey))
	builder.WriteString(fmt.Sprintf(" -t=%s", testPath))
	builder.WriteString(fmt.Sprintf(" -l=%s", resultPath))

	return builder.String()
}

func killCmdGen(loadTestKey string) string {
	grepRegex := fmt.Sprintf("'\\/bin\\/ApacheJMeter\\.jar.*%s'", loadTestKey)

	return fmt.Sprintf("kill -15 $(ps -ef | grep -E %s | awk '{print $2}')", grepRegex)
}

func readAndParseScript(scriptPath string) (string, error) {
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		log.Println("file doesn't exist on correct path")
		return "", err
	}

	return string(data), nil
}
