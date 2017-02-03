package main

import (
	"log"
	"fmt"
	"bufio"
	"os"
	"io"
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
			fmt.Print("\n")
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
