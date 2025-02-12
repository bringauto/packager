package bringauto_repository

import (
	"bringauto/modules/bringauto_testing"
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
	err := setupPackages()
	if err != nil {
		panic(fmt.Sprintf("can't setup packages - %s", err))
	}
	code := m.Run()
	err = bringauto_testing.DeletePackageFiles()
	if err != nil {
		panic(fmt.Sprintf("can't delete package files - %s", err))
	}
	os.Exit(code)
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

	err = repo.CopyToRepository(pack1, bringauto_testing.Pack1Name)
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

	err = repo.CopyToRepository(pack1, bringauto_testing.Pack2Name)
	if err != nil {
		t.Errorf("CopyToRepository failed - %s", err)
	}

	err = repo.CopyToRepository(pack2, bringauto_testing.Pack2Name)
	if err != nil {
		t.Errorf("CopyToRepository failed - %s", err)
	}

	err = repo.CopyToRepository(pack3, bringauto_testing.Pack3Name)
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

	err = repo.CopyToRepository(pack1, bringauto_testing.Pack1Name)
	if err != nil {
		t.Errorf("CopyToRepository failed - %s", err)
	}
	err = repo.CommitAllChanges()
	if err != nil {
		t.Errorf("can't commit changes - %s", err)
	}

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

	err = repo.CopyToRepository(pack1, bringauto_testing.Pack1Name)
	if err != nil {
		t.Errorf("CopyToRepository failed - %s", err)
	}
	err = repo.RestoreAllChanges()
	if err != nil {
		t.Errorf("can't restore changes - %s", err)
	}

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
		t.Error("git status not empty after RestoreAllChanges")
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
	err := bringauto_testing.SetupPackageFiles()
	if err != nil {
		return err
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
