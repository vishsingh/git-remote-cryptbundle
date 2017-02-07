package main

import (
	"io"
	"strings"
	"os"
	"fmt"
	"regexp"
)

type Remote interface {
	Lock() error
	Unlock() error

	GetBundles() (bundleNames []string, err error)
	PushBundle(encryptedBundleStream io.Reader) (newBundleName string, err error)
}

type fsRemote struct {
	path string
}

func (r *fsRemote) Lock() error {
	if r.path == "" {
		return fmt.Errorf("fsRemote has blank path")
	}

	return os.Mkdir(r.path + "/lock", 0)
}

func (r *fsRemote) Unlock() error {
	if r.path == "" {
		return fmt.Errorf("fsRemote has blank path")
	}

	return os.Remove(r.path + "/lock")
}

var bundleNameRegex = regexp.MustCompile(`^[0-9]{8}\.bundle\.gpg$`)

func (r *fsRemote) GetBundles() ([]string, error) {
	fd, err := os.Open(r.path)
	if err != nil {
		return nil, err
	}

	names, err := fd.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	bundleNames := []string{}
	for _, name := range names {
		if bundleNameRegex.MatchString(name) {
			bundleNames = append(bundleNames, name)
		}
	}

	return bundleNames, nil
}

func (r *fsRemote) PushBundle(ebs io.Reader) (string, error) {
	return "", fmt.Errorf("fsRemote: PushBundle unimplemented")
}

func parseRemoteUrl(remoteUrl string) (Remote, error) {
	if strings.HasPrefix(remoteUrl, "/") {
		return &fsRemote {
			path: remoteUrl,
		}, nil
	}

	return nil, fmt.Errorf("unable to parse remote url %q", remoteUrl)
}
