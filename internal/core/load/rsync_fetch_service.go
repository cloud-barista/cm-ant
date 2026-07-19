package load

import (
	"fmt"
	"sort"
	"strings"
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
	StepRec                        *stepRecorder // FR-MA2-PERF-007-08 (nil = no recording)
	Finished                       chan struct{} // closed when fetchData returns, after the final rsync (BAR-1413; nil = not awaited)
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
	if f.Finished != nil {
		defer close(f.Finished)
	}
	ticker := time.NewTicker(defaultFetchIntervalSec * time.Second)
	defer ticker.Stop()

	done := f.LoadTestDone
	for {
		select {
		case <-ticker.C:
			if !f.isRunning() {
				f.setFetchRunning(true)
				if _, err := rsyncFiles(f); err != nil {
					log.Error().Msgf("error while fetching data from rsync %s", err.Error())
				}
				f.setFetchRunning(false)
			}
		case <-done:
			// FR-MA2-PERF-007-08: record the final result collection as its own step so a
			// rsync failure is visible (previously it was only logged and the run was still
			// marked successed).
			f.StepRec.begin(constant.StepResultFetch, "Collecting results")
			retry := config.AppConfig.Load.Retry
			var fetchErr error
			var outcome fetchOutcome
			for retry > 0 {
				if !f.isRunning() {
					outcome, fetchErr = rsyncFiles(f)
					if fetchErr != nil {
						log.Error().Msgf("error while fetching data from rsync %s", fetchErr.Error())
					}
					break
				}
				time.Sleep(time.Duration(1<<4-retry) * time.Second)
				retry--
			}

			switch {
			case fetchErr != nil:
				f.StepRec.fail(constant.StepResultFetch, "Result collection failed", fmt.Sprintf("the load test finished but collecting the result files (rsync) failed: %v", fetchErr))
			case len(outcome.MissingMetrics) > 0:
				// The run stands on its own load figures. Say which charts will be empty and
				// why, so an empty chart is not read as "the load produced nothing".
				msg := fmt.Sprintf("Results collected, without %s metrics", strings.Join(outcome.MissingMetrics, ", "))
				f.StepRec.ok(constant.StepResultFetch, msg)
			default:
				f.StepRec.ok(constant.StepResultFetch, "Results collected")
			}

			return
		}
	}
}

// fetchOutcome says which result files made it across. The distinction matters: the load
// figures come from the main file, while the cpu/disk/memory/network files are metrics
// gathered alongside. Losing a metric file costs a chart; losing the main file costs the run.
type fetchOutcome struct {
	MissingMetrics []string // "cpu", "network", ... — collected but not retrievable
}

// rsyncFiles pulls the result files from the generator. It fails only when the main result is
// missing; metric files that did not arrive are reported instead, so a run whose load figures
// are intact is not thrown away for want of a network chart (BAR-1552).
//
// Metric files also tend to land later than the main one, so what looks missing now often
// arrives on a later pass — hence the retries here and the periodic fetch above.
func rsyncFiles(f *fetchDataParam) (fetchOutcome, error) {
	var outcome fetchOutcome
	loadTestKey := f.LoadTestKey
	installLocation := f.InstallLocation
	loadGeneratorInstallPath := f.InstallPath

	var wg sync.WaitGroup
	resultsPrefix := []string{""}

	if f.CollectAdditionalSystemMetrics {
		resultsPrefix = append(resultsPrefix, "_cpu", "_disk", "_memory", "_network")
	}

	type fileResult struct {
		prefix string
		err    error
	}
	resultChan := make(chan fileResult, len(resultsPrefix))

	resultFolderPath := utils.JoinRootPathWith("/result/" + loadTestKey)

	err := utils.CreateFolderIfNotExist(utils.JoinRootPathWith("/result"))
	if err != nil {
		return outcome, err
	}

	err = utils.CreateFolderIfNotExist(resultFolderPath)
	if err != nil {
		return outcome, err
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

			log.Info().Msgf("cmd for rsync: %s", cmd)
			err := utils.InlineCmd(cmd)

			if err == nil {
				resultChan <- fileResult{prefix: prefix}
				return // Success, exit the retry loop
			}

			log.Error().Msg(fmt.Sprintf("Error during rsync attempt %d for %s: %v", attempt, fileName, err))

			// Wait before retrying
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				retryDelay *= 2 // Exponential backoff
			}
		}

		resultChan <- fileResult{prefix: prefix, err: fmt.Errorf("failed to rsync %s after %d attempts", fileName, maxRetries)}
	}

	if installLocation == constant.Local || installLocation == constant.Remote {
		for _, p := range resultsPrefix {
			wg.Add(1)
			go rsync(p)
		}
	}

	wg.Wait()
	close(resultChan)

	var mainErr error
	for r := range resultChan {
		if r.err == nil {
			continue
		}
		if r.prefix == "" {
			mainErr = r.err // the load figures themselves
			continue
		}
		outcome.MissingMetrics = append(outcome.MissingMetrics, strings.TrimPrefix(r.prefix, "_"))
	}
	sort.Strings(outcome.MissingMetrics)

	return outcome, mainErr
}
