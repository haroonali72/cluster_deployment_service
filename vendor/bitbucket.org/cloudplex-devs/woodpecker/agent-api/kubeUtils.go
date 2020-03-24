package agent_api

import (
	"bufio"
	"bytes"
	"log"
	"os/exec"
)

func execKubectl(in *ExecKubectlRequest) (*ExecKubectlResponse, error) {

	command := "/usr/local/bin/kubectl"
	var args = in.Args
	stdout, stderr, err := runCommand(command, args, "/")

	resp := ExecKubectlResponse{
		Stdout: []string{stdout.String()},
		Stderr: []string{stderr.String()},
		Status: "command executed successfully",
	}

	if err != nil {
		resp.Status = "error in command execution"
	}

	return &resp, err
}

func runCommand(command string, cmdArgs []string, dir string) (bytes.Buffer, bytes.Buffer, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd := exec.Command(command, cmdArgs...)

	cmd.Dir = dir
	//cmd.Stdin = stdin
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	err := cmd.Run()

	//if err != nil {
	//	return stdout, stderr, err
	//}

	return stdout, stderr, err
}

func execKubectlStream(in *ExecKubectlRequest, stream AgentServer_ExecKubectlStreamServer) error {

	command := "/usr/local/bin/kubectl"
	scanner, err := runCommandStream(command, in.Args)
	if err != nil {
		return err
	}

	for scanner.Scan() {
		resp := ExecKubectlResponse{}
		line := scanner.Text()
		log.Println(line)
		resp.Stdout = []string{line}
		if err := stream.Send(&resp); err != nil {
			log.Println("5", err)
			return err
		}
	}

	return nil

}

func runCommandStream(command string, cmdArgs []string) (*bufio.Scanner, error) {

	cmd := exec.Command(command, cmdArgs...)
	standardOut, err1 := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	err := cmd.Start()

	if err != nil {
		return nil, err
	}

	if err1 != nil {
		return nil, err1
	}

	stdOut := bufio.NewScanner(standardOut)

	return stdOut, nil
	//for stdOut.Scan() {
	//	resp := Response{}
	//	line := stdOut.Text()
	//	log.Println(line)
	//	resp.Stdout =  []string{line}
	//	if err := stream.Send(&resp); err != nil {
	//		log.Println("5",err)
	//		return err
	//	}
	//}
	//
	//err = cmd.Wait()
	//if err != nil {
	//	log.Println("4",err)
	//	return err
	//}

	//return nil
}
