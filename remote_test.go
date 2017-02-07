package main

import (
	"testing"
	"os"
	"io/ioutil"
)

func stringsEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestFsRemoteGetBundles(t *testing.T) {
	var err error

	workDir, err := ioutil.TempDir("", "getbundletest")
	if err != nil {
		t.Fatalf("failed to create temp dir: %s", err.Error())
	}
	defer os.RemoveAll(workDir)

	r := &fsRemote {
		path: workDir,
	}

	expectBundles := func (expectedBundles ...string) {
		bundles, err := r.GetBundles()
		if err != nil {
			t.Error(err)
			return
		}

		if !stringsEqual(bundles, expectedBundles) {
			t.Errorf("%#v != %#v", bundles, expectedBundles)
		}
	}

	expectBundles()

	if err := r.Lock(); err != nil {
		t.Error(err)
		return
	}

	expectBundles()

	if err := r.Unlock(); err != nil {
		t.Error(err)
		return
	}

	expectBundles()

	touch := func (name string) error {
		newfd, err := os.Create(workDir + "/" + name)
		if err != nil {
			return err
		}
		newfd.Close()
		return nil
	}

	err = touch("00000000.bundle.gpg")
	if err != nil {
		t.Error(err)
		return
	}

	expectBundles("00000000.bundle.gpg")

	err = touch("0000000.bundle.gpg")
	if err != nil {
		t.Error(err)
		return
	}

	expectBundles("00000000.bundle.gpg")
}
