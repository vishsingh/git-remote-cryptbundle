package main

import (
	"io"
	"strings"
	"os"
	"fmt"
	"regexp"
	"strconv"
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
const bundleNameIndexLength = 8 // must match the regex
const bundleNameFormatString = "%08d.bundle.gpg" // must match the regex

func newBundleFilename(existingBundles []string) string {
	newIndex := 0
	for _, b := range existingBundles {
		if len(b) < bundleNameIndexLength {
			continue
		}

		numericPart := b[0:bundleNameIndexLength]

		n, err := strconv.Atoi(numericPart)
		if err != nil {
			continue
		}

		if n < newIndex {
			continue
		}

		newIndex = n + 1
	}

	return fmt.Sprintf(bundleNameFormatString, newIndex)
}

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
	// determine new filename based on existing bundles
	existingBundles, err := r.GetBundles()
	if err != nil {
		return "", err
	}

	desiredFilename := newBundleFilename(existingBundles)

	// exclusively create a file under r.path
	desiredFilepath := r.path + "/" + desiredFilename
	temporaryFilepath := desiredFilepath + ".new"

	const openflags = os.O_WRONLY | os.O_APPEND | os.O_CREATE | os.O_EXCL
	f, err := os.OpenFile(temporaryFilepath, openflags, 0600)
	if err != nil {
		return "", err
	}

	defer f.Close()
	if _, err := io.Copy(f, ebs); err != nil {
		return "", err
	}

	// atomically rename to desired filename
	if err := os.Rename(temporaryFilepath, desiredFilepath); err != nil {
		return "", err
	}

	// clear write bit
	if err := f.Chmod(0400); err != nil {
		return "", err
	}

	if err := f.Close(); err != nil {
		return "", err
	}

	return desiredFilename, nil
}

func parseRemoteUrl(remoteUrl string) (Remote, error) {
	if strings.HasPrefix(remoteUrl, "/") {
		return &fsRemote {
			path: remoteUrl,
		}, nil
	}

	return nil, fmt.Errorf("unable to parse remote url %q", remoteUrl)
}
