package agent_api

import (
	"errors"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// Server represents the gRPC server
type AgentServer struct {
	Name string
}

// SayHello generates response to a Ping request
func (s *AgentServer) SayHello(ctx context.Context, in *PingMessage) (*PingMessage, error) {
	log.Printf("Receive message %s", in.Greeting)
	return &PingMessage{Greeting: s.Name}, nil
}

func (s *AgentServer) CreateFile(ctx context.Context, in *CreateFileRequest) (*FileResponse, error) {
	log.Printf("CreateFile called")

	resp := FileResponse{Error: []string{}}
	for _, file := range in.Files {
		filePath := filepath.Join(file.Path, file.Name)
		createFile(filePath)
		err := writeFile(file.Data, filePath)
		log.Println("writing", filePath)
		if err != nil {
			resp.Error = append(resp.Error, err.Error())
		} else {
			log.Println("successfully wrote", filePath)
		}
	}

	if len(resp.Error) > 0 {
		return &FileResponse{Status: "error in creating all files"}, errors.New("error in creating all files")
	}

	return &FileResponse{Status: "successfully created all files"}, nil

}

func (s *AgentServer) DeleteFile(ctx context.Context, in *CreateFileRequest) (*FileResponse, error) {
	log.Printf("DeleteFile file called")
	resp := FileResponse{Error: []string{}}
	for _, file := range in.Files {
		filePath := filepath.Join(file.Path, file.Name)

		err := deleteFile(filePath)
		log.Println("deleting", filePath)
		if err != nil {
			log.Println("error deleting", filePath)
			resp.Error = append(resp.Error, err.Error())
		} else {
			log.Println("successfully deleted", filePath)
		}
	}

	if len(resp.Error) > 0 {
		return &FileResponse{Status: "error in deleting all files"}, nil
	}

	return &FileResponse{Status: "successfully deleted all files"}, nil
}

func (s *AgentServer) ExecKubectl(ctx context.Context, in *ExecKubectlRequest) (*ExecKubectlResponse, error) {
	log.Printf("ExecKubectl file called")

	resp, err := execKubectl(in)

	if err != nil {
		out := errors.New(strings.Join(resp.Stderr, ","))
		return resp, out
	}

	return resp, err
}

func (s *AgentServer) ExecKubectlStream(in *ExecKubectlRequest, stream AgentServer_ExecKubectlStreamServer) error {
	log.Printf("ExecKubectlStream  called")

	err := execKubectlStream(in, stream)

	return err
}

func (s *AgentServer) ExecHttp(ctx context.Context, in *HttpRequest) (*HttpResponse, error) {
	log.Printf("ExecHttp file called")

	data, code, header, err := httpCaller(in)
	//httpResp, err := httpCaller(in)

	resp := HttpResponse{}
	if err != nil {
		resp.Error = err.Error()
	}

	resp.ResponseCode = int32(code)
	resp.Body = data
	resp.Headers = getGrpcHeader(header)

	return &resp, nil
}

func getGrpcHeader(src http.Header) []*Headers {
	var dst []*Headers
	for k, vv := range src {
		for _, v := range vv {
			header := Headers{
				Key:   k,
				Value: v,
			}
			dst = append(dst, &header)
		}
	}
	return dst
}

func (s *AgentServer) ExecHttpStream(in *HttpRequest, stream AgentServer_ExecHttpStreamServer) error {

	return nil
}

func (s *AgentServer) RemoteSSHStream(in *SSHRequest, stream AgentServer_RemoteSSHStreamServer) error {
	log.Printf("RemoteSSHStream  called")
	err := execSSHStream(in, stream)
	return err
}

func (s *AgentServer) RemoteSCPStream(in *SCPRequest, stream AgentServer_RemoteSCPStreamServer) error {

	log.Printf("RemoteSSHStream  called")
	err := execSCPStream(in, stream)
	return err
}

func (s *AgentServer) RemoteSSH(ctx context.Context, request *SSHRequestUnary) (*SSHResponse, error) {
	log.Printf("RemoteSSH  called")
	resp, err := execSSHUnary(request)
	var errOut error
	if err != nil {
		errOut = errors.New(resp.Stderr)
	}
	return resp, errOut
}

func (s *AgentServer) RemoteSCP(ctx context.Context, request *SCPRequestUnary) (*SCPResponse, error) {

	log.Printf("RemoteSSH called")
	resp, err := execSCPUnary(request)
	//var errOut error
	//if err != nil {
	//	errOut = errors.New(resp.Error)
	//}
	return resp, err
}

func (s *AgentServer) Chunker(in *HttpRequest, stream AgentServer_ChunkerServer) error {
	return nil
}
