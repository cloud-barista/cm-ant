package load

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloud-barista/cm-ant/internal/utils"
)

// getAddAuthorizedKeyCommand returns a command to add the authorized key.
func getAddAuthorizedKeyCommand(pkName, pubkName string) (string, error) {
	pubKeyPath, _, err := validateKeyPair(pkName, pubkName)
	if err != nil {
		return "", err
	}

	pub, err := utils.ReadToString(pubKeyPath)
	if err != nil {
		return "", err
	}

	addAuthorizedKeyScript := utils.JoinRootPathWith("/script/add-authorized-key.sh")

	addAuthorizedKeyCommand, err := utils.ReadToString(addAuthorizedKeyScript)
	if err != nil {
		return "", err
	}

	addAuthorizedKeyCommand = strings.Replace(addAuthorizedKeyCommand, `PUBLIC_KEY=""`, fmt.Sprintf(`PUBLIC_KEY="%s"`, pub), 1)
	return addAuthorizedKeyCommand, nil
}

// validateKeyPair checks and generates SSH key pair if it doesn't exist.
func validateKeyPair(pkName, pubkName string) (string, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}

	privKeyPath := fmt.Sprintf("%s/.ssh/%s", homeDir, pkName)
	pubKeyPath := fmt.Sprintf("%s/.ssh/%s", homeDir, pubkName)

	err = utils.CreateFolderIfNotExist(fmt.Sprintf("%s/.ssh", homeDir))
	if err != nil {
		return pubKeyPath, privKeyPath, err
	}

	exist := utils.ExistCheck(privKeyPath)
	if !exist {
		err := utils.GenerateSSHKeyPair(4096, privKeyPath, pubKeyPath)
		if err != nil {
			return pubKeyPath, privKeyPath, err
		}
	}
	return pubKeyPath, privKeyPath, nil
}
