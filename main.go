package main

import (
	"fmt"
	"github.com/yext/yerrors"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	fmt.Println("SSH/SFTP Playground")

	// init logger
	logger, err := zap.NewDevelopment()

	username, ok := os.LookupEnv("REMOTE_USER")
	if !ok {
		logger.Error(yerrors.New("missing environment variable").Error())
		return
	}

	password, ok := os.LookupEnv("PASSWORD")
	if !ok {
		logger.Error(yerrors.New("missing environment variable").Error())
		return
	}

	remoteHost, ok := os.LookupEnv("REMOTE_HOST")
	if !ok {
		logger.Error(yerrors.New("missing environment variable").Error())
		return
	}

	portNumber := 22
	cmd := "hostname"

	privateKeyFile, ok := os.LookupEnv("PRIVATE_KEY_FILE_PATH")
	if !ok {
		logger.Error(yerrors.New("missing environment variable").Error())
		return
	}

	knownHost, ok := os.LookupEnv("KNOWN_HOST_SHA")
	if !ok {
		logger.Error(yerrors.New("missing environment variable").Error())
		return
	}

	_, _, _, _, _, err = ssh.ParseKnownHosts([]byte(knownHost))
	if err != nil {
		logger.Error(yerrors.Wrap(err).Error())
		return
	}

	fileReader, err := os.Open(privateKeyFile)
	if err != nil {
		logger.Error(yerrors.Wrap(err).Error())
		return
	}

	privKeyData, err := io.ReadAll(fileReader)
	if err != nil {
		logger.Error(yerrors.Wrap(err).Error())
		return
	}

	privKey, err := ssh.ParsePrivateKeyWithPassphrase(privKeyData, []byte(password))
	if err != nil {
		logger.Error(yerrors.Wrap(err).Error())
		return
	}
	log.Printf("found private key %s", privKey)
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privKey),
		},
		//HostKeyCallback: ssh.FixedHostKey(pubKey),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		BannerCallback:  nil,
		ClientVersion:   "",
		//HostKeyAlgorithms: []string{
		//	ssh.KeyAlgoRSASHA512,
		//},
		Timeout: 10 * time.Second,
	}

	serverAddress := fmt.Sprintf("%s:%d", remoteHost, portNumber)
	log.Printf("Connecting to %s | Configs %v", serverAddress, config)
	conn, err := ssh.Dial("tcp", serverAddress, config)
	if err != nil {
		logger.Error(yerrors.Wrap(err).Error())
		return
	}

	sess, err := conn.NewSession()
	if err != nil {
		logger.Error(yerrors.Wrap(err).Error())
		return
	}

	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	err = sess.Run(cmd)
	if err != nil {
		return
	}

	err = sess.Close()
	if err != nil {
		return
	}
}
