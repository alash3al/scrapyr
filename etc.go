package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func cmd(timeoutDuration time.Duration, argv []string, output *string) error {
	stdout, stderr := bytes.NewBufferString(""), bytes.NewBufferString("")

	c := exec.Command(argv[0], argv[1:]...)
	c.Dir = config.Scrapy.ProjectDir
	c.Stderr = stderr
	c.Stdout = stdout

	timeoutReached := false

	if timeoutDuration > 0 {
		time.AfterFunc(timeoutDuration, func() {
			c.Process.Signal(os.Kill)
			c.Process.Signal(os.Kill)

			timeoutReached = true
		})
	}

	if err := c.Run(); err != nil {
		return err
	}

	if stderr.Len() > 0 {
		return errors.New(stderr.String())
	}

	if output != nil {
		*output = stdout.String()
	}

	if timeoutReached {
		return fmt.Errorf("the execution time exceeded (%s)", timeoutDuration.String())
	}

	return nil
}

func catchErr(err error) {
	log.Println(err)
}
