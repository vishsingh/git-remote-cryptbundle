package main

import (
	"io"
	//"strings"
	//"os"
	"fmt"
	//"regexp"
	//"strconv"
)

type sshfsRemote struct {
}

func (*sshfsRemote) errNotImpl(method string) error {
	return fmt.Errorf("sshfsRemote %s(): Not implemented", method)
}

func (r *sshfsRemote) Lock() error {
	return r.errNotImpl("Lock")
}

func (r *sshfsRemote) Unlock() error {
	return r.errNotImpl("Unlock")
}

func (r *sshfsRemote) GetBundles() ([]string, error) {
	return nil, r.errNotImpl("GetBundles")
}

func (r *sshfsRemote) PushBundle(io.Reader) (string, error) {
	return "", r.errNotImpl("PushBundle")
}

func (r *sshfsRemote) CommitBundle() error {
	return r.errNotImpl("CommitBundle")
}
