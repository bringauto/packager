package bringauto_repository

import (
	"bringauto/modules/bringauto_package"
	"bringauto/modules/bringauto_prerequisites"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const (
	RepoName = "repo"
	ZipExtension = ".zip"

	pack1Name = "pack1"
	pack2Name = "pack2"
	pack3Name = "pack3"
	pack1FileName = pack1Name + "_file"
	pack2FileName = pack2Name + "_file"
	pack3FileName = pack3Name + "_file"
)

var defaultPlatformString bringauto_package.PlatformString
var pack1 bringauto_package.Package
var pack2 bringauto_package.Package
var pack3 bringauto_package.Package

func TestMain(m *testing.M) {
	stringExplicit := bringauto_package.PlatformStringExplicit {
		DistroName: "distro",
		DistroRelease: "1.0",
		Machine: "machine",
	}

	defaultPlatformString = bringauto_package.PlatformString{
		Mode: bringauto_package.ModeExplicit,
		String: stringExplicit,
	}
	setupPackages()
	m.Run()
	deletePackages()
}

func TestDirDoesNotExists(t *testing.T) {
	repo := GitLFSRepository {
		GitRepoPath: RepoName,
	}
	err := bringauto_prerequisites.Initialize(&repo)
	if err == nil {
		t.Fail()
	}
}

func TestDirIsNotGitRepo(t *testing.T) {
	err := os.MkdirAll(RepoName, 0755)
	if err != nil {
		t.Fatalf("can't create repo directory - %s", err)
	}

	repo := GitLFSRepository {
		GitRepoPath: RepoName,
	}
	err = bringauto_prerequisites.Initialize(&repo)
	if err == nil {
		t.Fail()
	}

	err = os.RemoveAll(RepoName)
	if err != nil {
		t.Fatalf("can't delete repo directory - %s", err)
	}
}

func TestCreatePackagePath(t *testing.T) {
	repo, err := initGitRepo()
	if err != nil {
		t.Fatalf("can't initialize Git repository or struct - %s", err)
	}

	packPath := repo.CreatePackagePath(pack1)
	expectedPackPath := filepath.Join(
		RepoName,
		pack1.PlatformString.String.DistroName,
		pack1.PlatformString.String.DistroRelease,
		pack1.PlatformString.String.Machine,
		pack1.Name,
	)

	if packPath != expectedPackPath {
		t.Fail()
	}

	err = deleteGitRepo()
	if err != nil {
		t.Fatalf("can't delete Git repository - %s", err)
	}
}

func TestCopyToRepositoryOnePackage(t *testing.T) {
	repo, err := initGitRepo()
	if err != nil {
		t.Fatalf("can't initialize Git repository or struct - %s", err)
	}

	err = repo.CopyToRepository(pack1, pack1Name)
	if err != nil {
		t.Errorf("CopyToRepository failed - %s", err)
	}

	packFilePath := filepath.Join(repo.CreatePackagePath(pack1), pack1.GetFullPackageName() + ZipExtension)
	_, err = os.ReadFile(packFilePath)
	if os.IsNotExist(err) {
		t.Fail()
	}

	err = deleteGitRepo()
	if err != nil {
		t.Fatalf("can't delete Git repository - %s", err)
	}
}

func TestCopyToRepositoryMultiplePackages(t *testing.T) {
	repo, err := initGitRepo()
	if err != nil {
		t.Fatalf("can't initialize Git repository or struct - %s", err)
	}

	err = repo.CopyToRepository(pack1, pack2Name)
	if err != nil {
		t.Errorf("CopyToRepository failed - %s", err)
	}

	err = repo.CopyToRepository(pack2, pack2Name)
	if err != nil {
		t.Errorf("CopyToRepository failed - %s", err)
	}

	err = repo.CopyToRepository(pack3, pack3Name)
	if err != nil {
		t.Errorf("CopyToRepository failed - %s", err)
	}

	pack1FilePath := filepath.Join(repo.CreatePackagePath(pack1), pack1.GetFullPackageName() + ZipExtension)
	_, err = os.ReadFile(pack1FilePath)
	if os.IsNotExist(err) {
		t.Fail()
	}

	pack2FilePath := filepath.Join(repo.CreatePackagePath(pack2), pack2.GetFullPackageName() + ZipExtension)
	_, err = os.ReadFile(pack2FilePath)
	if os.IsNotExist(err) {
		t.Fail()
	}

	pack3FilePath := filepath.Join(repo.CreatePackagePath(pack3), pack3.GetFullPackageName() + ZipExtension)
	_, err = os.ReadFile(pack3FilePath)
	if os.IsNotExist(err) {
		t.Fail()
	}

	err = deleteGitRepo()
	if err != nil {
		t.Fatalf("can't delete Git repository - %s", err)
	}
}

func TestCommitAllChanges(t *testing.T) {
	repo, err := initGitRepo()
	if err != nil {
		t.Fatalf("can't initialize Git repository or struct - %s", err)
	}

	err = repo.CopyToRepository(pack1, pack1Name)
	if err != nil {
		t.Errorf("CopyToRepository failed - %s", err)
	}
	repo.CommitAllChanges()

	err = os.Chdir(RepoName)
	if err != nil {
		t.Fatal("can't change directory")
	}

	cmd := exec.Command("git", "status", "-s")
	stdout, err := cmd.Output()
	if err != nil {
		t.Errorf("git status failed - %s", err)
	}
	if len(stdout) > 0 {
		t.Error("git status not empty after CommitAllChanges")
	}

	cmd = exec.Command("git", "log")
	_, err = cmd.Output()
	if err != nil {
		t.Error("no commit added")
	}

	err = os.Chdir("../")
	if err != nil {
		t.Fatal("can't change directory")
	}

	err = deleteGitRepo()
	if err != nil {
		t.Fatalf("can't delete Git repository - %s", err)
	}
}

func TestRestoreAllChanges(t *testing.T) {
	repo, err := initGitRepo()
	if err != nil {
		t.Fatalf("can't initialize Git repository or struct - %s", err)
	}

	err = repo.CopyToRepository(pack1, pack1Name)
	if err != nil {
		t.Errorf("CopyToRepository failed - %s", err)
	}
	repo.RestoreAllChanges()

	err = os.Chdir(RepoName)
	if err != nil {
		t.Fatal("can't change directory")
	}

	cmd := exec.Command("git", "status", "-s")
	stdout, err := cmd.Output()
	if err != nil {
		t.Errorf("git status failed - %s", err)
	}
	if len(stdout) > 0 {
		t.Error("git status not empty after CommitAllChanges")
	}

	cmd = exec.Command("git", "log")
	_, err = cmd.Output()
	if err == nil {
		t.Error("some commit added")
	}

	err = os.Chdir("../")
	if err != nil {
		t.Fatal("can't change directory")
	}

	err = deleteGitRepo()
	if err != nil {
		t.Fatalf("can't delete Git repository - %s", err)
	}
}

func initGitRepo() (GitLFSRepository, error) {
	err := os.MkdirAll(RepoName, 0755)
	if err != nil {
		return GitLFSRepository{}, err
	}
	err = os.Chdir(RepoName)
	if err != nil {
		return GitLFSRepository{}, err
	}

	cmd := exec.Command("git", "init")
	_, err = cmd.Output()
	if err != nil {
		return GitLFSRepository{}, err
	}

	err = os.Chdir("../")
	if err != nil {
		return GitLFSRepository{}, err
	}

	repo := GitLFSRepository {
		GitRepoPath: RepoName,
	}
	err = bringauto_prerequisites.Initialize(&repo)

	return repo, err
}

func deleteGitRepo() error {
	return os.RemoveAll(RepoName)
}

func setupPackages() error {
	err := os.Mkdir(pack1Name, 0755)
	if err != nil {
		return fmt.Errorf("failed to create a directory - %s", err)
	}
	err = os.Mkdir(pack2Name, 0755)
	if err != nil {
		return fmt.Errorf("failed to create a directory - %s", err)
	}
	err = os.Mkdir(pack3Name, 0755)
	if err != nil {
		return fmt.Errorf("failed to create a directory - %s", err)
	}

	file1, err := os.Create(filepath.Join(pack1Name, pack1FileName))
	if err != nil {
		return fmt.Errorf("failed to create a file - %s", err)
	}
	defer file1.Close()
	file2, err := os.Create(filepath.Join(pack2Name, pack2FileName))
	if err != nil {
		return fmt.Errorf("failed to create a file - %s", err)
	}
	defer file2.Close()
	file3, err := os.Create(filepath.Join(pack3Name, pack3FileName))
	if err != nil {
		return fmt.Errorf("failed to create a file - %s", err)
	}
	defer file3.Close()

	_, err = file1.WriteString("file1 content")
	if err != nil {
		return fmt.Errorf("failed to write to file - %s", err)
	}
	_, err = file2.WriteString("file2 content")
	if err != nil {
		return fmt.Errorf("failed to write to file - %s", err)
	}
	_, err = file3.WriteString("file3 content")
	if err != nil {
		return fmt.Errorf("failed to write to file - %s", err)
	}

	pack1 = bringauto_package.Package{
		Name: "pack1",
		VersionTag: "1.0",
		PlatformString: defaultPlatformString,
		IsDevLib: false,
		IsLibrary: false,
		IsDebug: false,
	}

	pack2 = bringauto_package.Package{
		Name: "pack2",
		VersionTag: "1.0",
		PlatformString: defaultPlatformString,
		IsDevLib: true,
		IsLibrary: true,
		IsDebug: false,
	}

	pack3 = bringauto_package.Package{
		Name: "pack3",
		VersionTag: "1.0",
		PlatformString: defaultPlatformString,
		IsDevLib: false,
		IsLibrary: true,
		IsDebug: false,
	}

	return nil
}

func deletePackages() error {
	err := os.RemoveAll(pack1Name)
	if err != nil {
		return err
	}
	err = os.RemoveAll(pack2Name)
	if err != nil {
		return err
	}
	err = os.RemoveAll(pack3Name)
	if err != nil {
		return err
	}
	return nil
}
