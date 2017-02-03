package main

import (
	"log"
	"fmt"
	"bufio"
	"os"
	"io"
	"strings"
)

func doIt() error {
	r := bufio.NewReader(os.Stdin)

	for {
		line, err := r.ReadString('\n')

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		line = line[0:len(line)-1]

		if line == "" {
			continue
		}

		if line == "capabilities" {
			fmt.Print("push\n")
			fmt.Print("\n")
		} else if line == "list for-push" {
			return fmt.Errorf("unimplemented")
		} else if strings.HasPrefix(line, "push ") {
			return fmt.Errorf("push command unimplemented")
		} else {
			return fmt.Errorf("unknown command: %s", line)
		}
	}

	return nil
}

func main() {
	err := doIt()

	if err != nil {
		log.Fatal(err)
	}
}
