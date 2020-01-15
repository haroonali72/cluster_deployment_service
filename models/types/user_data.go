package types

type UserData struct {
	WriteFile []WriteFile `yaml:"write_files,omitempty"`
	RunCmd    [][]string  `yaml:"runcmd,omitempty"`
}

type WriteFile struct {
	Path       string `yaml:"path,omitempty"`
	Owner      string `yaml:"owner,omitempty"`
	Permission string `yaml:"permissions,omitempty"`
	Encoding   string `yaml:"encoding,omitempty"`
	Contents   string `yaml:"content,omitempty"`
}

type Data struct {
	Agent  string     `json:"agent_binary"`
	Config ConfigFile `json:"config_file"`
}
type ConfigFile struct {
	AgentId             string `json:"agent_id"`
	AgentManagerAddress string `json:"agent_manager_address"`
	ClientCert          string `json:"client_cert"`
	ClientKey           string `json:"client_key"`
}