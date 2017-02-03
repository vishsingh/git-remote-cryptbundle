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

type pushCommand struct {
	force bool
	src string
	dst string
}

func parsePushCommand(p string) *pushCommand {
	return nil
}

func handlePushCommand(c *config, pc *pushCommand) error {
	return fmt.Errorf("unable to execute push command %#v", pc)
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
