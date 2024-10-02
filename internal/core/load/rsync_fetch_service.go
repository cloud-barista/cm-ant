package load

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/utils"
)

type fetchDataParam struct {
	LoadTestDone    <-chan bool
	LoadTestKey     string
	InstallLocation constant.InstallLocation
	InstallPath     string
	PublicKeyName   string
	PrivateKeyName  string
	Username        string
	PublicIp        string
	Port            string
	AgentInstalled  bool
	fetchMx         sync.Mutex
	fetchRunning    bool
	Home            string
}

func (f *fetchDataParam) setFetchRunning(running bool) {
	f.fetchMx.Lock()
	defer f.fetchMx.Unlock()
	f.fetchRunning = running
}

func (f *fetchDataParam) isRunning() bool {
	f.fetchMx.Lock()
	defer f.fetchMx.Unlock()
	return f.fetchRunning
}

const (
	defaultFetchIntervalSec = 30
)

func (l *LoadService) fetchData(f *fetchDataParam) {
	ticker := time.NewTicker(defaultFetchIntervalSec * time.Second)
	defer ticker.Stop()

	done := f.LoadTestDone
	for {
		select {
		case <-ticker.C:
			if !f.isRunning() {
				f.setFetchRunning(true)
				if err := rsyncFiles(f); err != nil {
					log.Println(err)
				}
				f.setFetchRunning(false)
			}
		case <-done:
			retry := 3
			for retry > 0 {
				if !f.isRunning() {
					if err := rsyncFiles(f); err != nil {
						log.Println(err)
					}
					break
				}
				time.Sleep(time.Duration(1<<4-retry) * time.Second)
				retry--
			}

			return
		}
	}
}

func rsyncFiles(f *fetchDataParam) error {
	loadTestKey := f.LoadTestKey
	installLocation := f.InstallLocation
	loadGeneratorInstallPath := f.InstallPath

	var wg sync.WaitGroup
	resultsPrefix := []string{""}

	if f.AgentInstalled {
		resultsPrefix = append(resultsPrefix, "_cpu", "_disk", "_memory", "_network")
	}

	errorChan := make(chan error, len(resultsPrefix))

	resultFolderPath := utils.JoinRootPathWith("/result/" + loadTestKey)

	err := utils.CreateFolderIfNotExist(utils.JoinRootPathWith("/result"))
	if err != nil {
		return err
	}

	err = utils.CreateFolderIfNotExist(resultFolderPath)
	if err != nil {
		return err
	}

	if installLocation == constant.Local {
		for _, p := range resultsPrefix {
			wg.Add(1)
			go func(prefix string) {
				defer wg.Done()
				fileName := fmt.Sprintf("%s%s_result.csv", loadTestKey, prefix)
				fromFilePath := fmt.Sprintf("%s/result/%s", loadGeneratorInstallPath, fileName)
				toFilePath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)

				cmd := fmt.Sprintf(`rsync -avz %s %s`, fromFilePath, toFilePath)
				utils.LogInfo("cmd for rsync: ", cmd)
				err := utils.InlineCmd(cmd)
				errorChan <- err
			}(p)
		}
	} else if installLocation == constant.Remote {
		for _, p := range resultsPrefix {
			wg.Add(1)
			go func(prefix string) {
				defer wg.Done()
				fileName := fmt.Sprintf("%s%s_result.csv", loadTestKey, prefix)
				fromFilePath := fmt.Sprintf("%s/result/%s", loadGeneratorInstallPath, fileName)
				toFilePath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)

				cmd := fmt.Sprintf(`rsync -avz -e "ssh -i %s -o StrictHostKeyChecking=no" %s@%s:%s %s`,
					fmt.Sprintf("%s/.ssh/%s", f.Home, f.PrivateKeyName),
					f.Username,
					f.PublicIp,
					fromFilePath,
					toFilePath)

				utils.LogInfo("cmd for rsync: ", cmd)
				err := utils.InlineCmd(cmd)
				errorChan <- err
			}(p)
		}
	}

	wg.Wait()
	close(errorChan)

	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	return nil
}
