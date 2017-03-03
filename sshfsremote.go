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

	inner Remote
}

func (*sshfsRemote) errNotImpl(method string) error {
	return fmt.Errorf("sshfsRemote %s(): Not implemented", method)
}

func sshfsCmd(url string, mountPoint string) *exec.Cmd {
	mountCmd := exec.Command("sshfs", url, mountPoint)

	options := []string{
		"reconnect",
		"sshfs_sync",
		"no_readahead",
		"sync_readdir",
	}

	for _, option := range options {
		mountCmd.Args = append(mountCmd.Args, "-o", option)
	}

	return mountCmd
}

func (r *sshfsRemote) Init() error {
	// mkdir <mountpoint>
	// sshfs <r.url> <mountpoint> <mount-options>

	if r.mountPoint != "" || r.mounted {
		return fmt.Errorf("sshfsRemote already initialized")
	}

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

	mountCmd := sshfsCmd(r.url, mountPoint)

	out, err := mountCmd.CombinedOutput()
	if err != nil {
		r.removeMountPoint()
		return fmt.Errorf("failed to mount sshfs: %q", string(out))
	}

	r.mounted = true

	r.inner = &fsRemote {
		path: mountPoint,
	}

	return r.inner.Init()
}

func (r *sshfsRemote) removeMountPoint() error {
	if r.mountPoint != "" {
		if err := os.Remove(r.mountPoint); err != nil {
			return fmt.Errorf("failed to remove sshfs mountpoint: %v", err)
		}

		r.mountPoint = ""
	}

	return nil
}

func (r *sshfsRemote) Uninit() error {
	// fusermount -u <mountpoint>
	// rmdir <mountpoint>

	if r.inner != nil {
		if err := r.inner.Uninit(); err != nil {
			return err
		}

		r.inner = nil
	}

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

	return r.removeMountPoint()
}

func (r *sshfsRemote) Lock() error {
	return r.inner.Lock()
}

func (r *sshfsRemote) Unlock() error {
	return r.inner.Unlock()
}

func (r *sshfsRemote) GetBundles() ([]string, error) {
	return r.inner.GetBundles()
}

func (r *sshfsRemote) PushBundle(ebs io.Reader) (string, error) {
	return r.inner.PushBundle(ebs)
}

func (r *sshfsRemote) CommitBundle() error {
	return r.inner.CommitBundle()
}
