package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func AddToKnownHost(pemFilePath, publicIp, username string) error {
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

// GenerateSSHKeyPair generates an RSA SSH key pair of the specified bits.
func GenerateSSHKeyPair(bits int, privKeyPath, pubKeyPath string) error {
	reader := rand.Reader

	key, err := rsa.GenerateKey(reader, bits)
	if err != nil {
		return err
	}

	pub, err := ssh.NewPublicKey(key.Public())
	if err != nil {
		return err
	}
	err = savePrivateKey(key, privKeyPath)
	if err != nil {
		return nil
	}

	err = savePublicKey(pub, pubKeyPath)
	if err != nil {
		return err
	}

	return nil
}

// savePrivateKey saves the private key to a file.
func savePrivateKey(key *rsa.PrivateKey, filename string) error {
	// Marshal private key to PEM format
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}

	// Write to file
	err := os.WriteFile(filename, pem.EncodeToMemory(block), 0600)
	if err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}
	return nil
}

// savePublicKey saves the public key to a file.
func savePublicKey(key ssh.PublicKey, filename string) error {
	// Marshal public key to authorized_keys format
	keyBytes := ssh.MarshalAuthorizedKey(key)

	// Write to file
	err := os.WriteFile(filename, keyBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}
	return nil
}

func DownloadFile(client *ssh.Client, toFilePath string, fromFilePath string) error {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	localFile, err := os.Create(toFilePath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	remoteFile, err := sftpClient.Open(fromFilePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	stat, err := remoteFile.Stat()
	if err != nil {
		return err
	}

	if stat.Size() == 0 {
		return errors.New("file is empty")
	}

	_, err = io.Copy(localFile, remoteFile)

	if err != nil {
		if err != io.EOF {
			return err
		}
	}
	return nil
}

func GetClient(publicIp, port, username, privateKeyName string) (*ssh.Client, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	privateKeyPath := fmt.Sprintf("%s/.ssh/%s", home, privateKeyName)

	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", publicIp, port), sshConfig)

	if err != nil {
		return nil, err
	}

	return client, nil
}
