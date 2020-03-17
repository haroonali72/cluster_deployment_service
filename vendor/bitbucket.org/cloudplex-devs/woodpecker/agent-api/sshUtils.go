package agent_api

import (
	"bufio"
	"fmt"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

func execSSHStream(in *SSHRequest, stream AgentServer_RemoteSSHStreamServer) error {

	sr, err := NewScriptRunner(in.Server, in.User, in.Key, "", false)

	if err != nil {
		log.Println(err)
		return err
	}
	defer sr.Client.Close()

	for _, cmd := range in.Cmd {
		log.Println(cmd)
		stdOut, stdErr, err := sr.ExecuteCmd(cmd, sr.Client)
		if err != nil {
			log.Println(err)
			return err
		}
		resp := SSHResponse{Stdout: *stdOut, Stderr: *stdErr, Status: "", Command: cmd}
		if err := stream.Send(&resp); err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func execSCPStream(in *SCPRequest, stream AgentServer_RemoteSCPStreamServer) error {

	sr, err := NewScriptRunner(in.Server, in.User, in.Key, "", false)

	if err != nil {
		log.Println(err)
		return err
	}

	defer sr.Client.Close()

	for _, file := range in.Files {

		resp := SCPResponse{}
		err := sr.ExecuteScp(file.Name, file.Path, file.Data)
		resp.Name = file.Name
		resp.Path = file.Path

		if err != nil {
			resp.Status = fmt.Sprintf("failure in creating file %s", file.Name)
			resp.Error = err.Error()
		} else {
			resp.Status = fmt.Sprintf("file with name %s created successfully", file.Name)
		}
		if err := stream.Send(&resp); err != nil {
			log.Println(err)
			return err
		}

	}

	return nil
}

func execSSHUnary(in *SSHRequestUnary) (*SSHResponse, error) {

	log.Printf("executing command %s", in.Cmd)
	sr, err := NewScriptRunner(in.Server, in.User, in.Key, "", false)

	if err != nil {
		log.Println(err)
		return &SSHResponse{}, err
	}
	defer sr.Client.Close()

	stdOut, stdErr, err := sr.ExecuteCmd(in.Cmd, sr.Client)
	if err != nil {
		log.Println(err)
		return &SSHResponse{}, err
	}
	resp := SSHResponse{Stdout: *stdOut, Stderr: *stdErr, Status: "", Command: in.Cmd}

	return &resp, nil
}

func execSCPUnary(in *SCPRequestUnary) (*SCPResponse, error) {

	log.Printf("creating file %s at %s", in.Files.Name, in.Files.Path)
	sr, err := NewScriptRunner(in.Server, in.User, in.Key, "", false)

	if err != nil {
		log.Println(err)
		return &SCPResponse{}, err
	}

	defer sr.Client.Close()

	//for _, file := range in.Files {

	resp := &SCPResponse{}
	err = sr.ExecuteScp(in.Files.Name, in.Files.Path, in.Files.Data)
	resp.Name = in.Files.Name
	resp.Path = in.Files.Path

	if err != nil {
		resp.Status = fmt.Sprintf("failure in creating file %s", in.Files.Name)
		resp.Error = err.Error()
	} else {
		resp.Status = fmt.Sprintf("file with name %s created successfully", in.Files.Name)
	}

	return resp, err
}

type ScriptRunner struct {
	Host   string
	Config *ssh.ClientConfig
	Client *ssh.Client
}

func NewScriptRunner(host, user, keyfile, password string, isPassword bool) (*ScriptRunner, error) {

	//https://github.com/golang/crypto
	sr := ScriptRunner{}

	if isPassword {
		config, err := getClientConfigWithPassword(user, password)
		if err != nil {
			return &sr, err
		}
		config.KeyExchanges = append(config.KeyExchanges, "curve25519-sha256@libssh.org")
		sr.Config = config

	} else {
		//, "ecdh-sha2-nistp256", "ecdh-sha2-nistp384", "ecdh-sha2-nistp521", "diffie-hellman-group14-sha1", "diffie-hellman-group1-sha1"
		config, err := getClientConfigWithKey(user, keyfile)
		if err != nil {
			return &sr, err
		}
		config.KeyExchanges = append(config.KeyExchanges, "curve25519-sha256@libssh.org")
		sr.Config = config
	}

	sr.Host = host

	err := sr.Dial()
	if err != nil {
		log.Println(err)
		return &sr, err
	}
	return &sr, nil
}

func (sr *ScriptRunner) Dial() error {
	client, err := ssh.Dial("tcp", sr.Host+":22", sr.Config)
	sr.Client = client
	return err
}

func (sr *ScriptRunner) ExecuteCmd(cmd string, conn *ssh.Client) (*string, *string, error) {

	session, err := conn.NewSession()
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	//defer session.Close()

	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()

	err = session.Start(cmd)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	standard_out := io.Reader(stdout)
	standard_err := io.Reader(stderr)

	inp := bufio.NewScanner(standard_out)
	errs := bufio.NewScanner(standard_err)

	resp := ""
	for inp.Scan() {
		line := inp.Text()
		resp = resp + line + "\n"
		log.Println(line)
		//utils.Info.Println("cmd => " + line)
	}

	errResp := ""

	for errs.Scan() {
		line := errs.Text()
		errResp = errResp + line + "\n"
		log.Println(line)

	}

	return &resp, &errResp, nil

}

func (sr *ScriptRunner) ExecuteCmdStream(cmd string) (*bufio.Scanner, error) {

	conn, err := ssh.Dial("tcp", sr.Host+":22", sr.Config)
	if err != nil {
		return nil, err
	}
	session, err := conn.NewSession()
	if err != nil {
		return nil, err
	}
	//defer session.Close()

	session.Stderr = session.Stdout

	stdout, err3 := session.StdoutPipe()

	if err3 != nil {
		return nil, err3
	}

	_ = session.Start(cmd)

	standardOut := io.Reader(stdout)

	inp := bufio.NewScanner(standardOut)

	return inp, nil

}

func (sr *ScriptRunner) ExecuteScp(scriptName, path, content string) error {

	f, _ := ioutil.TempFile("", "")
	fmt.Fprintln(f, content)
	f.Close()
	defer os.Remove(f.Name())

	conn, err := ssh.Dial("tcp", sr.Host+":22", sr.Config)

	if err != nil {
		return err
	}
	session, err := conn.NewSession()

	if err != nil {
		return err
	}
	defer session.Close()
	err = scp.CopyPath(f.Name(), path+"/"+scriptName, session)

	return err
}

//func (sr *ScriptRunner) ExecuteScpFile(filename string) error {
//
//	conn, err := ssh.Dial("tcp", sr.Host+":22", sr.Config)
//	if err != nil {
//		//utils.SendError(err.Error(), logData, constants.LogBackend|constants.LogBackend)
//		return err
//	}
//	session, err := conn.NewSession()
//
//	if err != nil {
//		//utils.SendError(err.Error(), logData, constants.LogBackend|constants.LogBackend)
//		return err
//	}
//	defer session.Close()
//	//dest := scriptName
//	err = scp.CopyPath(constants.BasePath+"/"+sr.ProjectId+"/"+filename, "/tmp/"+sr.ScriptPath+"/"+filename, session)
//	return err
//}

func getClientConfigWithPassword(user, password string) (*ssh.ClientConfig, error) {

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}

	return sshConfig, nil
}

func getClientConfigWithKey(user, key string) (*ssh.ClientConfig, error) {

	signer, err := makeSigner(key)
	if err != nil {
		return nil, err
	}
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Second * 5,
	}

	return sshConfig, nil
}

func makeSigner(key string) (signer ssh.Signer, err error) {
	//basePath := constants.BasePath
	//utils.Info.Println("key", basePath+keyname)
	//contents, errf := ioutil.ReadFile(basePath + keyname)
	//if errf != nil {
	//	fmt.Println(errf)
	//	return nil, errors.New(keyname + " file does not exist")
	//}
	buf := []byte(key)
	signer, err = ssh.ParsePrivateKey(buf)
	return signer, err
}
