package config

type Server struct {
	*Config
	Host          *string `toml:"host,omitempty"`
	Port          *int64  `toml:"port,omitempty"`
	Operator      *Entity `toml:"operator,omitempty"`
	Account       *Entity `toml:"account,omitempty"`
	SystemAccount *Entity `toml:"system_account,omitempty"`
}

func NewDefaultServer() (*Server, error) {
	conf, err := NewDefaultConfig()
	if err != nil {
		return nil, err
	}

	host := "0.0.0.0"
	port := int64(443)
	return &Server{
		Config: conf,
		Host:   &host,
		Port:   &port,
	}, nil
}

func ParseServerFile(file string) (*Server, error) {
	var conf *Server
	err := ParseFile(file, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
