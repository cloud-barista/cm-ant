package utils

import (
	"log"
	"os"
	"os/exec"
)

func InlineCmd(cmdStr string) error {
	cmd := exec.Command("bash", "-c", cmdStr)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("error while execute bash call,", err)
		log.Println(string(out))
		return err
	}

	log.Println(string(out))
	return nil
}

func Script(shellScriptPath string, envs []string, args ...string) error {
	cmd := exec.Command(shellScriptPath, args...)

	if len(envs) > 0 {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, envs...)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("error while execute bash call,", err)
		log.Println(string(out))
		return err
	}

	log.Println(string(out))
	return nil
}

func InlineCmdAsync(cmdStr string) error {
	cmd := exec.Command("bash", "-c", cmdStr)

	var err error

	if err = cmd.Start(); err != nil {
		log.Println("error while async bash call:", err)
		return err
	}

	return nil
}
