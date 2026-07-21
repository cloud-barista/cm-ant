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

			// Wait for the metric files rather than taking one look and moving on.
			//
			// Every chart on the results screen is drawn from these csv files, and they are
			// written while the run is still going, so the metric ones routinely land after
			// the main result. There is no way to ask for them again afterwards, so a run
			// that gives up here loses its charts for good.
			//
			// Waiting stops when the set of missing files stops shrinking: files still
			// arriving means more are coming, while nothing new across a whole round means
			// nothing more is coming.
			outcome, fetchErr := l.collectResults(f)

			switch {
			case fetchErr != nil:
				f.StepRec.fail(constant.StepResultFetch, "Result collection failed",
					fmt.Sprintf("the load test finished but the main result file could not be collected from the generator (%s): %v",
						f.PublicIp, fetchErr))
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

// collectResults keeps fetching until the files stop arriving.
//
// It returns as soon as everything is in. Otherwise it keeps going while progress is being
// made - each round that brings a new file earns another round - and stops once a whole round
// adds nothing, or when the overall deadline passes.
//
// Each result file is recorded as its own sub-step (BAR: result-fetch granularity), so a
// collection that sits for a long time shows exactly which file it is still waiting on rather
// than one opaque "Collecting results". A 30s run whose results take far longer is almost
// always waiting on a specific metric csv here, and that file is now named as it waits.
func (l *LoadService) collectResults(f *fetchDataParam) (fetchOutcome, error) {
	deadline := time.Now().Add(resultCollectWindow)
	var outcome fetchOutcome
	var err error
	previousMissing := -1

	// Name every file we expect up front so the console draws the whole set from the first poll.
	f.beginResultFileSteps()

	for round := 1; ; round++ {
		if f.isRunning() {
			time.Sleep(resultCollectInterval)
			continue
		}

		outcome, err = rsyncFiles(f)
		if err != nil {
			log.Error().Msgf("error while fetching data from rsync %s", err.Error())
			f.recordStep(constant.SubFileResult, "failed", "Load result file not collected", err.Error())
			return outcome, err
		}
		// The main file is in (err == nil); flip it and any metric files that have arrived to ok,
		// and leave the rest waiting with the round they are on.
		f.recordResultFileProgress(outcome, round)

		if len(outcome.MissingMetrics) == 0 {
			return outcome, nil
		}

		madeProgress := previousMissing < 0 || len(outcome.MissingMetrics) < previousMissing
		previousMissing = len(outcome.MissingMetrics)

		if !madeProgress || time.Now().After(deadline) {
			log.Warn().Msgf("giving up on metric files after %d rounds: %v", round, outcome.MissingMetrics)
			// The load figures are in; the missing metric files only cost their charts.
			f.skipMissingMetricFiles(outcome)
			return outcome, nil
		}

		f.StepRec.progress(constant.StepResultFetch, round,
			fmt.Sprintf("Collecting results - still waiting for %s", strings.Join(outcome.MissingMetrics, ", ")),
			"metric files are written during the run and can arrive after the main result")
		time.Sleep(resultCollectInterval)
	}
}

// resultFilePrefixes lists the rsync prefixes this run collects: the main result, plus the
// metric files when additional system metrics were requested. It matches rsyncFiles exactly.
func (f *fetchDataParam) resultFilePrefixes() []string {
	prefixes := []string{""}
	if f.CollectAdditionalSystemMetrics {
		prefixes = append(prefixes, "_cpu", "_disk", "_memory", "_network")
	}
	return prefixes
}

// resultFileStep maps an rsync prefix to its sub-step name.
func resultFileStep(prefix string) constant.ExecutionStep {
	switch prefix {
	case "":
		return constant.SubFileResult
	case "_cpu":
		return constant.SubFileCpu
	case "_disk":
		return constant.SubFileDisk
	case "_memory":
		return constant.SubFileMemory
	case "_network":
		return constant.SubFileNetwork
	}
	return ""
}

// resultFileLabel is the human name for a result file, used in the step message.
func resultFileLabel(prefix string) string {
	switch prefix {
	case "":
		return "load result"
	case "_cpu":
		return "cpu metrics"
	case "_disk":
		return "disk metrics"
	case "_memory":
		return "memory metrics"
	case "_network":
		return "network metrics"
	}
	return strings.TrimPrefix(prefix, "_")
}

// recordStep is a nil-safe helper for the file sub-steps.
func (f *fetchDataParam) recordStep(name constant.ExecutionStep, status, message, detail string) {
	if f.StepRec == nil || name == "" {
		return
	}
	switch status {
	case "ok":
		f.StepRec.ok(name, message)
	case "failed":
		f.StepRec.fail(name, message, detail)
	case "skipped":
		f.StepRec.skip(name, message)
	default:
		f.StepRec.begin(name, message)
	}
}

// beginResultFileSteps marks every expected file as waiting, so the set is visible from the start.
func (f *fetchDataParam) beginResultFileSteps() {
	for _, p := range f.resultFilePrefixes() {
		f.recordStep(resultFileStep(p), "running", "Waiting for the "+resultFileLabel(p)+" file", "")
	}
}

// recordResultFileProgress flips the files that have arrived to ok and leaves the rest waiting.
func (f *fetchDataParam) recordResultFileProgress(outcome fetchOutcome, round int) {
	if f.StepRec == nil {
		return
	}
	missing := map[string]bool{}
	for _, m := range outcome.MissingMetrics {
		missing[m] = true
	}
	for _, p := range f.resultFilePrefixes() {
		metric := strings.TrimPrefix(p, "_")
		if p == "" {
			f.StepRec.ok(constant.SubFileResult, "Load result file collected")
			continue
		}
		if missing[metric] {
			f.StepRec.progress(resultFileStep(p), round,
				fmt.Sprintf("Waiting for the %s file (round %d)", resultFileLabel(p), round),
				"written on the generator during the run; can arrive after the load result")
		} else {
			f.StepRec.ok(resultFileStep(p), resultFileLabel(p)+" collected")
		}
	}
}

// skipMissingMetricFiles records the metric files that never arrived. They are skipped, not
// failed: the run stands on its load result and only loses those charts (BAR-1552).
func (f *fetchDataParam) skipMissingMetricFiles(outcome fetchOutcome) {
	for _, metric := range outcome.MissingMetrics {
		f.recordStep(resultFileStep("_"+metric), "skipped", resultFileLabel("_"+metric)+" not collected in time", "")
	}
}

const (
	resultCollectInterval = 10 * time.Second
	resultCollectWindow   = 10 * time.Minute
)

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
