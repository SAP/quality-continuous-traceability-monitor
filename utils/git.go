package utils

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/golang/glog"
)

// CommitAndPush to git repo
func CommitAndPush(repoPath string) {

	os.Chdir(repoPath)
	CallGit("add", ".")

	// Set identity
	CallGit("config", "user.email", "ctm@myCorp.com")
	CallGit("config", "user.name", "Continuous Traceability Monitor")

	now := time.Now().UTC().Format(time.RFC1123)
	CallGit("commit", "-m", "Update traceability report "+now)

	CallGit("push")

}

// CloneRepo from remote
func CloneRepo(cfg Config, repoURL, repoPath string) {

	if _, err := os.Stat(repoPath); err == nil {
		glog.Error("Unable to clone ", repoURL, " to ", repoPath, ". Directory already exists")
		panic(err)
	} else {
		if !os.IsNotExist(err) {
			glog.Error("Unable to clone ", repoURL, " to ", repoPath)
			panic(err)
		}
	}

	repoURLToken := addAccessTokenToURL(cfg, repoURL)

	err := CallGit("clone", repoURLToken, repoPath)
	if err != nil {
		glog.Fatal("Unable to clone ", repoURL, " to ", repoPath)
	}

}

// PullRepo from remote
func PullRepo(repoPath string) error {

	os.Chdir(repoPath)
	err := CallGit("pull")

	return err

}

// CloneOrPullRepo from remote
// If the given repo exists locally it will try to pull (no force!) the repo (in order to update it)
// if it doesn't exist, it will clone the repo.
// If resetOnFailedPull is set to true it will call a 'git reset --hard' if the preceding pull command failed
func CloneOrPullRepo(cfg Config, repoURL, repoPath string, resetOnFailedPull bool) {

	if Exists(repoPath) {
		err := PullRepo(repoPath)
		if err != nil && resetOnFailedPull { // Pull command failed. Let's delete the repo and clone again
			CallGit("reset", "--hard")
		}
	} else {
		CloneRepo(cfg, repoURL, repoPath)
	}

}

// CallGit and pass given params as git command line client parameters
func CallGit(params ...string) error {

	cmd := exec.Command("git", params...)
	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		glog.Error(err)
	}

	return err

}

// GetRepositoryNameFromURL only works on GitHub repos
func GetRepositoryNameFromURL(url string) string {

	repoName := url[strings.LastIndex(url, "/")+1:]
	repoName = strings.TrimSuffix(repoName, ".git")

	return repoName

}

// GetRepositoryHTTPSUrl for GitHub repo
func GetRepositoryHTTPSUrl(cfg Config, git Git) string {

	return cfg.Github.BaseURL + "/" + git.Organization + "/" + git.Repository
}

// GetRepositorySSHUrl for GitHub repo
func GetRepositorySSHUrl(cfg Config, git Git) string {

	base := strings.Replace(cfg.Github.BaseURL, "https://", "git@", -1)
	base = strings.TrimSuffix(base, "/")

	return base + ":" + git.Organization + "/" + git.Repository + ".git"
}

// IsGitInstalled check whether git command line client is accessible on local machine
func IsGitInstalled() bool {

	cmd := exec.Command("git", "--version")
	err := cmd.Run()
	if err != nil {
		return false
	}

	return true
}

func addAccessTokenToURL(cfg Config, repoURL string) string {

	if cfg.Github.AccessToken != "" {
		repoURL = strings.Replace(repoURL, "https://", "https://"+cfg.Github.AccessToken+"@", 1)
	}

	return repoURL

}
