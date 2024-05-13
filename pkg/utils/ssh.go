package utils

import (
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"os"
	"path/filepath"
)

func AddToKnownHost(pemFileName, publicIp, username string) error {
	pemFilePath := filepath.Join(os.Getenv("HOME"), ".ssh", pemFileName)
	auth, err := goph.Key(pemFilePath, "")
	if err != nil {
		return err
	}

	client, err := goph.NewConn(&goph.Config{
		User:     username,
		Addr:     publicIp,
		Port:     22,
		Auth:     auth,
		Callback: VerifyHost,
	})

	if err != nil {
		return err
	}

	defer client.Close()

	out, err := client.Run("uptime")

	if err != nil {
		return err
	}

	log.Println(string(out))

	return nil
}

func VerifyHost(host string, remote net.Addr, key ssh.PublicKey) error {

	hostFound, err := goph.CheckKnownHost(host, remote, key, "")

	if hostFound && err != nil {

		return err
	}

	if hostFound && err == nil {

		return nil
	}

	return goph.AddKnownHost(host, remote, key, "")
}
