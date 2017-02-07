package main

import (
	"io"
	"fmt"
)

type Remote interface {
	Lock() error
	Unlock() error

	GetBundles() (bundleNames []string, err error)
	PushBundle(encryptedBundleStream io.Reader) (newBundleName string, err error)
}

func parseRemoteUrl(remoteUrl string) (Remote, error) {
	return nil, fmt.Errorf("unable to parse remote url %q", remoteUrl)
}
