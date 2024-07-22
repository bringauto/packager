package bringauto_ssh

import (
	"bufio"
	"fmt"
	"github.com/pkg/sftp"
	"github.com/mholt/archiver/v3"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"time"
)

const (
	archiveName string = "install_arch.tar"
	archiveNameSep string = string(os.PathSeparator) + archiveName
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
func (sftpd *SFTP) DownloadDirectory() error {
	var err error

	tar := Tar{
		ArchiveName: archiveName,
		SourceDir: "/INSTALL",
	}

	shellEvaluator := ShellEvaluator{
		Commands: tar.ConstructCMDLine(),
		StdOut:   os.Stdout,
	}

	shellEvaluator.RunOverSSH(*sftpd.SSHCredentials)

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

	localArchivePath := sftpd.EmptyLocalDir + archiveNameSep

	fmt.Printf("%s Copying tar with sftp\n", time.Now())

	err = sftpd.copyFile(sftpClient, sftpd.RemoteDir + archiveNameSep, localArchivePath)
	if err != nil {
		return fmt.Errorf("cannot copy recursive %s", err)
	}

	fmt.Printf("%s File copied. Unarchiving tar.\n", time.Now())

	tarArchive := archiver.Tar{
		OverwriteExisting: false,
		MkdirAll: false,
		ImplicitTopLevelFolder: false,
		ContinueOnError: true,
	}

	tarArchive.Unarchive(localArchivePath, sftpd.EmptyLocalDir)

	err = os.Remove(localArchivePath)
	if err != nil {
		return fmt.Errorf("cannot remove local dir %s", err)
	}

	return nil
}

func (sftpd *SFTP) copyFile(sftpClient *sftp.Client, remoteFile string, localDir string) error {
	var err error
	_, err = sftpClient.Lstat(remoteFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("requested remote file %s does not exist", remoteFile)
	}
	normalizedLocalDir, _ := normalizePath(localDir)

	remotePathStat, err := sftpClient.Lstat(remoteFile)
	if err != nil {
		return fmt.Errorf("cannot get Lstat if remote %s", remoteFile)
	}
	sourceFile, err := sftpClient.Open(remoteFile)
	if err != nil {
		return fmt.Errorf("cannot open file for read - %s,%s", remoteFile, err)
	}
	destFile, err := os.OpenFile(normalizedLocalDir, os.O_RDWR|os.O_CREATE, remotePathStat.Mode().Perm())
	if err != nil {
		return err
	}

	copyIOFile(sourceFile, destFile)

	return nil
}

// deprecated, reason: very slow with big trees, new approach: create tar and and copy single archive with copyFile()
func (sftpd *SFTP) copyRecursive(sftpClient *sftp.Client, remoteDir string, localDir string) error {
	var err error
	_, err = sftpClient.Lstat(sftpd.RemoteDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("requested remote file %s does not exist", sftpd.RemoteDir)
	}
	normalizedRemoteDir, _ := normalizePath(remoteDir)
	normalizedLocalDir, _ := normalizePath(localDir)

	allDone := make(chan bool, 2000)
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

			copyIOFile(sourceFile, destFile)
		}()

	}
	//Problem: this system copy files directory by directory in linear manner
	// just stupid wait mechanism
	for i := 0; i < fileCount; i++ {
		<-allDone
	}

	return nil
}

func copyIOFile(sourceFile *sftp.File, destFile *os.File) {
	sourceFileBuff := bufio.NewReaderSize(sourceFile, 1024*1024)
	destFileBuff := bufio.NewWriterSize(destFile, 1027*1024)

	var err error
	_, err = io.Copy(destFileBuff, sourceFileBuff)
	if err != nil {
		panic(fmt.Errorf("cannot copy remote IO files"))
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
}

// normalizePath
func normalizePath(p string) (string, error) {
	regexp, regexpErr := regexp.CompilePOSIX("[/]{2,}")
	if regexpErr != nil {
		return "", fmt.Errorf("sftp cannot compile regex - %s", regexpErr)
	}
	normalizePath := regexp.ReplaceAllString(p, "/")
	return normalizePath, nil
}
