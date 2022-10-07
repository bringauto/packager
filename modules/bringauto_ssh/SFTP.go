package bringauto_ssh

import (
	"bufio"
	"fmt"
	"github.com/pkg/sftp"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
)

type SFTP struct {
	// Path to a directory on the remote machine
	RemoteDir string
	// Empty, existing local directory where the RemoteDir will be copy
	EmptyLocalDir  string
	SSHCredentials *SSHCredentials
}

// DownloadDirectory
// Download directory from RemoteDir to EmptyLocalDir.
// EmptyLocalDir must be empty!
// Function returns error in case of problem or nil if succeeded.
//
func (sftpd *SFTP) DownloadDirectory() error {
	var err error

	sshSession := SSHSession{}
	err = sshSession.LoginMultipleAttempts(*sftpd.SSHCredentials)
	if err != nil {
		return fmt.Errorf("SFTP DownloadDirectory error - %s", err)
	}

	sftpClient, err := sftp.NewClient(sshSession.sshClient,
		sftp.MaxConcurrentRequestsPerFile(64),
		sftp.UseConcurrentReads(true),
		sftp.UseFstat(true),
		sftp.MaxPacket(1<<15),
	)
	if err != nil {
		return fmt.Errorf("SFTP DownloadDirectory problem - %s", err)
	}
	defer sftpClient.Close()

	if _, err = os.Stat(sftpd.EmptyLocalDir); os.IsNotExist(err) {
		return fmt.Errorf("EmptyLocalDir '%s' does not exist", sftpd.EmptyLocalDir)
	}

	localPathDirContent, _ := ioutil.ReadDir(sftpd.EmptyLocalDir)
	localPathDirIsNotEmpty := len(localPathDirContent) != 0
	if localPathDirIsNotEmpty {
		return fmt.Errorf("local directory '%s' is not empty", sftpd.EmptyLocalDir)
	}

	err = sftpd.copyRecursive(sftpClient, sftpd.RemoteDir, sftpd.EmptyLocalDir)
	if err != nil {
		return fmt.Errorf("cannot copy recursive %s", err)
	}

	return nil
}

func (sftpd *SFTP) copyRecursive(sftpClient *sftp.Client, remoteDir string, localDir string) error {
	var err error
	_, err = sftpClient.Lstat(sftpd.RemoteDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("requested remote file %s does not exist", sftpd.RemoteDir)
	}
	normalizedRemoteDir, _ := normalizePath(remoteDir)
	normalizedLocalDir, _ := normalizePath(localDir)

	allDone := make(chan bool)
	fileCount := 0

	walk := sftpClient.Walk(normalizedRemoteDir)
	for walk.Step() {
		if walk.Err() != nil {
			continue
		}
		remotePath, _ := normalizePath(walk.Path())
		if normalizedRemoteDir == remotePath {
			continue
		}
		relativeRemotePath := remotePath[len(normalizedRemoteDir):]
		absoluteLocalPath := path.Join(normalizedLocalDir, relativeRemotePath)
		remotePathStat, err := sftpClient.Lstat(remotePath)
		if err != nil {
			return fmt.Errorf("cannot get Lstat if remote %s", normalizedRemoteDir)
		}

		if remotePathStat.IsDir() {
			err = os.MkdirAll(absoluteLocalPath, remotePathStat.Mode().Perm())
			if err != nil {
				return fmt.Errorf("cannot create local directory - %s", err)
			}
			err = sftpd.copyRecursive(sftpClient, remotePath, absoluteLocalPath)
			if err != nil {
				return fmt.Errorf("sftp copy - %s", err)
			}
			continue
		}

		sourceFile, err := sftpClient.Open(remotePath)
		if err != nil {
			return fmt.Errorf("cannot open file for read - %s,%s", remotePath, err)
		}
		destFile, err := os.OpenFile(absoluteLocalPath, os.O_RDWR|os.O_CREATE, remotePathStat.Mode().Perm())
		if err != nil {
			return err
		}

		fileCount += 1
		go func() {
			defer func() { allDone <- true }()
			sourceFileBuff := bufio.NewReaderSize(sourceFile, 1024*1024*2)
			destFileBuff := bufio.NewWriterSize(destFile, 1027*1024*2)

			_, err = io.Copy(destFileBuff, sourceFileBuff)
			if err != nil {
				panic(fmt.Errorf("cannot copy remote file %s to dest file %s", remotePath, absoluteLocalPath))
			}

			_ = destFileBuff.Flush()

			err = destFile.Close()
			if err != nil {
				panic(fmt.Errorf("cannot close destFile: %s", err))
			}
			err = sourceFile.Close()
			if err != nil {
				panic(fmt.Errorf("cannot close sourceFile: %s", err))
			}
		}()

	}
	// just stupid wait mechanism
	for i := 0; i < fileCount; i++ {
		<-allDone
	}

	return nil
}

// normalizePath
//
func normalizePath(p string) (string, error) {
	regexp, regexpErr := regexp.CompilePOSIX("[/]{2,}")
	if regexpErr != nil {
		return "", fmt.Errorf("sftp cannot compile regex - %s", regexpErr)
	}
	normalizePath := regexp.ReplaceAllString(p, "/")
	return normalizePath, nil
}
