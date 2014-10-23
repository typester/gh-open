package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
)

type Remote struct {
	Name, Url, Type string
}

func DetectRemote(dir string) ([]*Remote, error) {
	git := exec.Command("git", "remote", "-v")
	git.Dir = dir

	stdout, err := git.StdoutPipe()
	if err != nil {
		return nil, err
	}
	git.Stderr = os.Stderr

	err = git.Start()
	if err != nil {
		return nil, err
	}

	remotes := make([]*Remote, 0)

	reader := bufio.NewReader(stdout)
	re := regexp.MustCompile(`^(\S+)\s*(\S+)\s*\((\S+)\)`)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		m := re.FindStringSubmatch(line)
		if m != nil {
			r := &Remote{m[1], m[2], m[3]}
			remotes = append(remotes, r)
		}
	}

	if err := git.Wait(); err != nil {
		return nil, err
	}

	return remotes, nil
}

// Convert git remote url to github URL
//
// Supported types:
//   - git@github.com:username/repo.git
//   - https://github.com/username/repo.git
//   - git://github.com/username/repo.git
//   - git@bitbucket.org:username/repo.git //same github
//   - https://username@bitbucket.org/username/repo.git //diffrent github
func MangleURL(url string) (string, error) {
	ssh_re := regexp.MustCompile(`^git@(.*?):(.*?)/(.*?)\.git$`)
	ssh2_re := regexp.MustCompile(`^ssh://git@(.*?)/(.*?)/(.*?)\.git$`)
	https_re := regexp.MustCompile(`^https://(.*?)/(.*?)/(.*?)(?:.git)?$`)
	https2_re := regexp.MustCompile(`^https://.*@(.*?)/(.*?)/(.*?)(?:.git)?$`)
	git_re := regexp.MustCompile(`^git://(.*?)/(.*?)/(.*?).git$`)

	var matches []string

	if m := ssh_re.FindStringSubmatch(url); m != nil {
		matches = m
	} else if m := ssh2_re.FindStringSubmatch(url); m != nil {
		matches = m
	} else if m := https2_re.FindStringSubmatch(url); m != nil {
		matches = m
	} else if m := https_re.FindStringSubmatch(url); m != nil {
		matches = m
	} else if m := git_re.FindStringSubmatch(url); m != nil {
		matches = m
	} else {
		return "", fmt.Errorf("unsupported remote url: %s", url)
	}

	return CreateURL(matches[1], matches[2], matches[3])
}

func CreateURL(host, user, repo string) (string, error) {
	if host == "github.com" {
		return fmt.Sprintf("https://%s/%s/%s", host, user, repo), nil
	} else if host == "bitbucket.org" {
		return fmt.Sprintf("https://%s/%s/%s", host, user, repo), nil
	} else {
		return CreateURLByHubConfig(host, user, repo)
	}
}

func CreateURLByHubConfig(host, user, repo string) (string, error) {
	hub_host := GetConfig("hub.host", "")
	protocol := GetConfig("hub.protocol", "https")
	if hub_host != "" {
		if protocol != "https" && protocol != "http" {
			return "", fmt.Errorf("unsupported protocol: %s", protocol)
		}
		return fmt.Sprintf("%s://%s/%s/%s", protocol, host, user, repo), nil
	} else {
		return "", fmt.Errorf("invalid github (includes enterprise) or bitbucket host: %s", host)
	}
}

func GetConfig(name string, default_value string) string {
	value, err := exec.Command("git", "config", name).Output()
	if err != nil {
		return default_value
	}
	return string(value[:len(value)-1])
}
