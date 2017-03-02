package main

import (
	"io"
	//"strings"
	"os"
	"fmt"
	//"regexp"
	//"strconv"
	"os/exec"
)

type sshfsRemote struct {
	// The information required to mount the sshfs remote.
	// Will be the first argument passed to the 'sshfs' command.
	// Looks like: user@host:dir
	url string

	// Temporary dir that serves as SSHFS mountpoint.
	mountPoint string

	// True while SSHFS is mounted.
	mounted bool
}

func (*sshfsRemote) errNotImpl(method string) error {
	return fmt.Errorf("sshfsRemote %s(): Not implemented", method)
}

func (r *sshfsRemote) Init() error {
	// mkdir <mountpoint>
	// sshfs <r.url> <mountpoint> <mount-options>

	if r.url == "" {
		return fmt.Errorf("sshfsRemote given empty URL")
	}

	home := os.Getenv("HOME")
	if home == "" {
		return fmt.Errorf("can not find user's HOME dir")
	}

	// todo: allow user to customize
	mountPoint := home + "/.sshfs-mount"

	if err := os.Mkdir(mountPoint, 0700); err != nil {
		return fmt.Errorf("failed to create sshfs mountpoint: %v", err)
	}

	r.mountPoint = mountPoint

	mountCmd := exec.Command("sshfs",
		r.url,
		mountPoint) // todo: options

	out, err := mountCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to mount sshfs: %q", string(out))
	}

	r.mounted = true

	return nil
}

func (r *sshfsRemote) Uninit() error {
	// fusermount -u <mountpoint>
	// rmdir <mountpoint>

	if r.mounted {
		unmountCmd := exec.Command("fusermount",
			"-u",
			r.mountPoint)

		out, err := unmountCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to unmount sshfs: %q", string(out))
		}

		r.mounted = false
	}

	if r.mountPoint != "" {
		if err := os.Remove(r.mountPoint); err != nil {
			return fmt.Errorf("failed to remove sshfs mountpoint: %v", err)
		}

		r.mountPoint = ""
	}

	return nil
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
