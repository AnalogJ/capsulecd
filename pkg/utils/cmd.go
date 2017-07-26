package utils

import (
	"capsulecd/pkg/errors"
	"fmt"
	"os"
	"os/exec"
	"path"
)

//http://craigwickesser.com/2015/02/golang-cmd-with-custom-environment/
//http://www.ryanday.net/2012/10/01/installing-go-and-gopath/
//

func BashCmdExec(cmd string, workingDir string, environ []string, logPrefix string) error {
	return CmdExec("sh", []string{"-c", cmd}, workingDir, environ, logPrefix)
}

func CmdExec(cmdName string, cmdArgs []string, workingDir string,  environ []string, logPrefix string) error {

	if logPrefix == "" {
		logPrefix = " >> "
	} else {
		logPrefix = logPrefix + " | "
	}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if environ != nil{
		cmd.Env = environ
	}
	if workingDir != "" && path.IsAbs(workingDir) {
		cmd.Dir = workingDir
	} else if workingDir != "" {
		return errors.Custom("Working Directory must be an absolute path")
	}
	//cmdReader, err := cmd.StdoutPipe()
	//if err != nil {
	//	fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
	//	return err
	//}
	//
	//done := make(chan struct{})
	//
	//scanner := bufio.NewScanner(cmdReader)
	//go func() {
	//	for scanner.Scan() {
	//		fmt.Printf("%s%s\n", logPrefix, scanner.Text())
	//	}
	//	done <- struct{}{}
	//
	//}()

	err := cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		return err
	}

	//<-done

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		return err
	}
	return nil
}
