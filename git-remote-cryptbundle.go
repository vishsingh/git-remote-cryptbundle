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
	// Name of remote, passed as the first argument to a remote helper.
	remoteName string

	remoteUrl string
	remote Remote
	localGitDir string
}

func gitListRemotes(gitDir string) ([]string, error) {
	cmd := exec.Command("git",
		"--git-dir=" + gitDir,
		"remote")

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	outs := strings.TrimSpace(string(out))

	outlines := strings.Split(outs, "\n")

	return outlines, nil
}

func gitAssertRemote(gitDir string, remoteName string) error {
	remotes, err := gitListRemotes(gitDir)
	if err != nil {
		return err
	}

	for _, r := range remotes {
		if r == remoteName {
			return nil
		}
	}

	return fmt.Errorf("repo at %s does not contain a remote named %q", gitDir, remoteName)
}

func (c *config) Validate() error {
	fi, err := os.Stat(c.localGitDir)
	if err != nil {
		return fmt.Errorf("GIT_DIR not valid: %s", err.Error())
	}
	if !fi.IsDir() {
		return fmt.Errorf("GIT_DIR not actually a dir")
	}

	if err := gitAssertRemote(c.localGitDir, c.remoteName); err != nil {
		return fmt.Errorf("unable to verify existence of remote named %q: %s", c.remoteName, err.Error())
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

func localRefTrackingRemoteRef(c *config, remoteRef string) (string, error) {
	// refs/heads/master -> refs/remotes/{c.remoteName}/master

	const prefix = "refs/heads/"

	if !strings.HasPrefix(remoteRef, prefix) {
		return "", fmt.Errorf("unable to determine local ref tracking remote ref %q", remoteRef)
	}

	branchName := remoteRef[len(prefix):]

	ref := "refs/remotes/" + c.remoteName + "/" + branchName
	if _, err := evaluateRef(c.localGitDir, ref); err != nil {
		return "", fmt.Errorf("ref %s does not exist", ref)
	}

	return ref, nil
}

func evaluateRef(gitDir string, ref string) (string, error) {
	cmd := exec.Command("git",
		"--git-dir=" + gitDir,
		"rev-parse",
		"--verify",
		"--quiet",
		ref)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	outs := strings.TrimSpace(string(out))

	return outs, nil
}

func refsEqual(gitDir string, refA string, refB string) (bool, error) {
	refAEval, err := evaluateRef(gitDir, refA)
	if err != nil {
		return false, err
	}

	refBEval, err := evaluateRef(gitDir, refB)
	if err != nil {
		return false, err
	}

	return refAEval == refBEval, nil
}

// Perform the actual push, updating pc.dst on the destination with pc.src acquired from the local repo.
// todo: force only if pc.force is set
func handlePushCommand(c *config, pc *pushCommand) error {
	ctx, ctxDoneFunc := context.WithCancel(context.Background())
	defer ctxDoneFunc()

	// use the remote tracking branch, for now. later we will want to decrypt a refs list from the remote. (todo)

	nothingToDo := false

	var bundleRevList string
	if localRef, err := localRefTrackingRemoteRef(c, pc.dst); err == nil {
		if eq, _ := refsEqual(c.localGitDir, localRef, pc.src); eq {
			nothingToDo = true
		}

		bundleRevList = localRef + ".." + pc.src
	} else {
		bundleRevList = pc.src
		log.Printf("uploading full bundle as no local ref found that tracks remote ref %q\n", pc.dst)
	}

	if nothingToDo {
		log.Printf("remote tracking branch is up to date: nothing to do\n")
		return nil
	}

	bundleCmd := exec.CommandContext(ctx, "git",
		"--git-dir=" + c.localGitDir,
		"bundle",
		"create",
		"-",
		bundleRevList)

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

	if err := c.remote.CommitBundle(); err != nil {
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
		remoteName: args[1],
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
