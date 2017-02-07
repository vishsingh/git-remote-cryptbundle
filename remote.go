package main

import (
	"io"
	"strings"
	"os"
	"fmt"
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

func (r *fsRemote) GetBundles() ([]string, error) {
	return nil, fmt.Errorf("fsRemote: GetBundles unimplemented")
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
