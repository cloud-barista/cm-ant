package load

import (
	"fmt"

	"sync"
	"time"

	"github.com/cloud-barista/cm-ant/internal/config"
	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/utils"
	"github.com/rs/zerolog/log"
)

type fetchDataParam struct {
	LoadTestDone                   <-chan bool
	LoadTestKey                    string
	InstallLocation                constant.InstallLocation
	InstallPath                    string
	PublicKeyName                  string
	PrivateKeyName                 string
	Username                       string
	PublicIp                       string
	Port                           string
	CollectAdditionalSystemMetrics bool
	fetchMx                        sync.Mutex
	fetchRunning                   bool
	Home                           string
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
					log.Error().Msgf("error while fetching data from rsync %s", err.Error())
				}
				f.setFetchRunning(false)
			}
		case <-done:
			retry := config.AppConfig.Load.Retry
			for retry > 0 {
				if !f.isRunning() {
					if err := rsyncFiles(f); err != nil {
						log.Error().Msgf("error while fetching data from rsync %s", err.Error())
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

	if f.CollectAdditionalSystemMetrics {
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

	maxRetries := config.AppConfig.Load.Retry
	retryDelay := 2 * time.Second

	rsync := func(prefix string) {
		defer wg.Done()
		fileName := fmt.Sprintf("%s%s_result.csv", loadTestKey, prefix)
		fromFilePath := fmt.Sprintf("%s/result/%s", loadGeneratorInstallPath, fileName)
		toFilePath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)

		// Retry logic
		for attempt := 1; attempt <= maxRetries; attempt++ {
			var cmd string
			if installLocation == constant.Local {
				cmd = fmt.Sprintf(`rsync -avvz --no-whole-file %s %s`, fromFilePath, toFilePath)
			} else if installLocation == constant.Remote {
				cmd = fmt.Sprintf(`rsync -avvz -e "ssh -i %s -o StrictHostKeyChecking=no" %s@%s:%s %s`,
					fmt.Sprintf("%s/.ssh/%s", f.Home, f.PrivateKeyName),
					f.Username,
					f.PublicIp,
					fromFilePath,
					toFilePath)
			}

			utils.LogInfo("cmd for rsync: ", cmd)
			err := utils.InlineCmd(cmd)

			if err == nil {
				errorChan <- nil
				return // Success, exit the retry loop
			}

			utils.LogError(fmt.Sprintf("Error during rsync attempt %d for %s: %v", attempt, fileName, err))

			// Wait before retrying
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				retryDelay *= 2 // Exponential backoff
			}
		}

		errorChan <- fmt.Errorf("failed to rsync %s after %d attempts", fileName, maxRetries)
	}

	if installLocation == constant.Local || installLocation == constant.Remote {
		for _, p := range resultsPrefix {
			wg.Add(1)
			go rsync(p)
		}
	}

	wg.Wait()
	close(errorChan)

	// Collect errors from the error channel
	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	return nil
}
