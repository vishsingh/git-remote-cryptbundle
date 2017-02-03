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

		fmt.Printf("I have read a line: '%s'\n", line)
	}

	return nil
}

func main() {
	err := doIt()

	if err != nil {
		log.Fatal(err)
	}
}
