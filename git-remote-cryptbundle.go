package main

import (
	"log"
	"fmt"
	"bufio"
	"os"
	"os/exec"
	"context"
	"io"
	"strings"
)

type config struct {
	remoteUrl string
	remote Remote
	localGitDir string
}

func (c *config) Validate() error {
	fi, err := os.Stat(c.localGitDir)
	if err != nil {
		return fmt.Errorf("GIT_DIR not valid: %s", err.Error())
	}
	if !fi.IsDir() {
		return fmt.Errorf("GIT_DIR not actually a dir")
	}

	return nil
}

func (c *config) EvalRemote() error {
	remote, err := parseRemoteUrl(c.remoteUrl)
	if err != nil {
		return err
	}

	c.remote = remote

	return nil
}

type pushCommand struct {
	force bool
	src string
	dst string
}

func parsePushCommand(p string) *pushCommand {
	ret := new(pushCommand)

	if !strings.HasPrefix(p, "push ") {
		return nil
	}
	p = p[5:]

	if p == "" {
		return nil
	}

	if p[0] == '+' {
		ret.force = true
		p = p[1:]
	}

	fields := strings.SplitN(p, ":", 2)
	if len(fields) != 2 {
		return nil
	}

	ret.src = fields[0]
	ret.dst = fields[1]

	return ret
}

// Perform the actual push, updating pc.dst on the destination with pc.src acquired from the local repo.
// todo: force only if pc.force is set
func handlePushCommand(c *config, pc *pushCommand) error {
	ctx, ctxDoneFunc := context.WithCancel(context.Background())
	defer ctxDoneFunc()

	// todo: last-bundle
	bundleCmd := exec.CommandContext(ctx, "git",
		"--git-dir=" + c.localGitDir,
		"bundle",
		"create",
		"-",
		pc.src)

	bundleStream, err := bundleCmd.StdoutPipe()
	if err != nil {
		return err
	}

	bundleCmd.Stderr = os.Stderr

	if err := bundleCmd.Start(); err != nil {
		return err
	}

	// todo
	encryptionRecipient := "RockMan"

	encryptCmd := exec.CommandContext(ctx, "gpg",
		"--batch",
		"--encrypt",
		"--recipient", encryptionRecipient,
		"--cipher-algo", "AES256",
		"--force-mdc")

	encryptCmd.Stdin = bundleStream

	encryptedStream, err := encryptCmd.StdoutPipe()
	if err != nil {
		return err
	}

	encryptCmd.Stderr = os.Stderr

	if err := encryptCmd.Start(); err != nil {
		return err
	}

	if err := c.remote.Lock(); err != nil {
		return err
	}

	if _, err := c.remote.PushBundle(encryptedStream); err != nil {
		return err
	}

	if err := encryptCmd.Wait(); err != nil {
		return err
	}

	if err := bundleCmd.Wait(); err != nil {
		return err
	}

	// Only unlock if we get to the end with no errors.
	// Leaving the remote in a locked state is an indication that
	// something went wrong during the last push.
	if err := c.remote.Unlock(); err != nil {
		return err
	}

	return nil
}

func handlePush(c *config, pushCommands []string) {
	for _, p := range pushCommands {
		pc := parsePushCommand(p)
		if pc == nil {
			errStr := "unable to parse push command: " + p
			fmt.Printf("error ??? %q\n", errStr)

			continue
		}

		err := handlePushCommand(c, pc)
		if err != nil {
			fmt.Printf("error %s %q\n", pc.dst, err.Error())

			continue
		}

		fmt.Printf("ok %s\n", pc.dst)
	}

	fmt.Printf("\n")
}

// todo: check errors returned by fmt

func doIt(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("too few args")
	}

	c := &config {
		remoteUrl: args[2],
		localGitDir: os.Getenv("GIT_DIR"),
	}

	if err := c.Validate(); err != nil {
		return err
	}

	if err := c.EvalRemote(); err != nil {
		return err
	}

	log.Println("working with remote at URL:", c.remoteUrl)

	r := bufio.NewReader(os.Stdin)

	pushCmds := []string{}

	for {
		line, err := r.ReadString('\n')

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		line = line[0:len(line)-1]

		if line == "" {
			if len(pushCmds) > 0 {
				handlePush(c, pushCmds)
				pushCmds = []string{}
			}

			continue
		}

		if line == "capabilities" {
			fmt.Print("push\n")
			fmt.Print("\n")

			continue
		}

		if line == "list for-push" {
			// todo
			fmt.Print("\n")

			continue
		}

		if strings.HasPrefix(line, "push ") {
			pushCmds = append(pushCmds, line)

			continue
		}

		return fmt.Errorf("unknown command: %s", line)
	}

	return nil
}

func main() {
	err := doIt(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}
