package main

import (
	"log"
	"fmt"
	"bufio"
	"os"
	"io"
	"strings"
)

type config struct {
	remoteUrl string
}

func handlePush(c *config, pushCommand string) error {
	return fmt.Errorf("unable to perform '%s' to remote '%s'", pushCommand, c.remoteUrl)
}

// todo: check errors returned by fmt

func doIt(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("too few args")
	}

	c := &config {
		remoteUrl: args[2],
	}

	log.Println("working with remote at URL:", c.remoteUrl)

	r := bufio.NewReader(os.Stdin)

	type readState int
	const (
		stateDefault readState = iota
		stateReadPushCommands
	)
	s := stateDefault

	for {
		line, err := r.ReadString('\n')

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		line = line[0:len(line)-1]

		switch s {
		case stateDefault:
			if line == "" {
				continue
			}

			if line == "capabilities" {
				fmt.Print("push\n")
				fmt.Print("\n")
			} else if line == "list for-push" {
				// todo
				fmt.Print("\n")
			} else if strings.HasPrefix(line, "push ") {
				return handlePush(c, line)
			} else {
				return fmt.Errorf("unknown command: %s", line)
			}

		default:
			panic("unknown state encountered")
		}
	}

	return nil
}

func main() {
	err := doIt(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}
