package main

import (
	"io/ioutil"
	"os/exec"
	"regexp"
	"syscall"
	"testing"
)

func TestDetectRemote_NotGit(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error("failed to create tempdir:", err)
	}
	defer syscall.Rmdir(dir)

	remotes, err := DetectRemote(dir)

	if remotes != nil {
		t.Error("unexpected result:", remotes)
	}
	if err == nil {
		t.Error("error should be set")
	}

	re := regexp.MustCompile(`^exit status \d+`)
	if re.MatchString(err.Error()) == false {
		t.Error("unexpected error", err)
	}
}

func TestDetectRemote_NotFound(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error("failed to create tempdir:", err)
	}
	defer syscall.Rmdir(dir)

	remotes, err := DetectRemote(dir + "/not_found")

	if remotes != nil {
		t.Error("unexpected result:", remotes)
	}
	if err == nil {
		t.Error("error should be set")
	}

	re := regexp.MustCompile(`^chdir `)
	if re.MatchString(err.Error()) == false {
		t.Error("unexpected error", err)
	}
}

func TestDetectRemote(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error("failed to create tempdir:", err)
	}
	defer syscall.Rmdir(dir)

	git := exec.Command("git", "init")
	git.Dir = dir
	if err := git.Run(); err != nil {
		t.Error("failed to run git init:", err)
	}

	git = exec.Command("git", "remote", "add", "origin", "git@github.com:username/repo.git")
	git.Dir = dir
	if err := git.Run(); err != nil {
		t.Error("failed to run git init:", err)
	}

	remotes, err := DetectRemote(dir)

	if err != nil {
		t.Error("error should be nil:", err)
	}

	if len(remotes) < 1 {
		t.Error("unexpected remotes count")
	}

	if remotes[0].Name != "origin" {
		t.Error("unexpected remote name", remotes[0].Name)
	}
	if remotes[0].Url != "git@github.com:username/repo.git" {
		t.Error("unexpected remote url", remotes[0].Url)
	}
}

func TestMangleURL(t *testing.T) {
	expected := "https://github.com/username/repo"

	// ssh
	u, err := MangleURL("git@github.com:username/repo.git")
	if err != nil {
		t.Error("error should be nil:", err)
	}
	if u != expected {
		t.Error("unexpected url:", u)
	}

	// https
	u, err = MangleURL("https://github.com/username/repo.git")
	if err != nil {
		t.Error("error should be nil:", err)
	}
	if u != expected {
		t.Error("unexpected url:", u)
	}

	// git
	u, err = MangleURL("git://github.com/username/repo.git")
	if err != nil {
		t.Error("error should be nil:", err)
	}
	if u != expected {
		t.Error("unexpected url:", u)
	}

	// different host
	u, err = MangleURL("git@example.com:username/repo.git")
	if err == nil {
		t.Error("error should be set:", err)
	}
	if err.Error() != "invalid github host: example.com" {
		t.Error("unexpected error:", err)
	}

	// unsupported host
	u, err = MangleURL("git@example.com:repo.git")
	if err == nil {
		t.Error("error should be set:", err)
	}
	if err.Error() != "unsupported remote url: git@example.com:repo.git" {
		t.Error("unexpected error:", err)
	}
}

func TestCreateURL(t *testing.T) {
	gheHost := "ghe.exsample.com"
	gheProtocol := "http"

	ConfigSet("--remove", "gh-open")
	ConfigSet("gh-open.ghe-host", gheHost)
	ConfigSet("gh-open.ghe-protocol", gheProtocol)

	expected := "http://ghe.exsample.com/username/repo"

	u, err := CreateURL(gheHost, "username", "repo")

	ConfigSet("--remove", "gh-open")

	if err != nil {
		t.Error("error should be nil:", err)
	}
	if u != expected {
		t.Error("unexpected url:", u)
	}
}

func ConfigSet(key string, value string) {
	exec.Command("git", "config", "--global", key, value)
}
