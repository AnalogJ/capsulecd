package utils

import (
	"os/exec"
	"fmt"
	"os"
	"bufio"
)

func CmdExec(cmdName string, cmdArgs []string, workingDir string, logPrefix string) error {

	if(logPrefix == ""){
		logPrefix = logPrefix + " >> "
	} else {
		logPrefix = logPrefix + " | "
	}

	cmd := exec.Command(cmdName, cmdArgs...)
	if(workingDir != ""){
		cmd.Dir = workingDir
	}
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		return err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s%s\n", logPrefix, scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		return err
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		return err
	}
	return nil
}