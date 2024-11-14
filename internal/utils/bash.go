package utils

import (
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
)

func InlineCmd(cmdStr string) error {
	cmd := exec.Command("bash", "-c", cmdStr)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Msgf("error while execute bash call; %v", err)
		log.Error().Msg(string(out))
		return err
	}

	log.Info().Msgf(string(out))
	return nil
}

func Script(scriptPath string, envs []string, args ...string) error {
	cmd := exec.Command(scriptPath, args...)

	if len(envs) > 0 {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, envs...)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Msgf("error while execute bash call; %v", err)
		log.Error().Msgf(string(out))
		return err
	}

	log.Info().Msgf(string(out))
	return nil
}

func InlineCmdAsync(cmdStr string) error {
	cmd := exec.Command("bash", "-c", cmdStr)

	var err error

	if err = cmd.Start(); err != nil {
		log.Error().Msgf("error while async bash call; %v", err)
		return err
	}

	return nil
}
