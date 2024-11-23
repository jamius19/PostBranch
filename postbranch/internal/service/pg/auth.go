package pg

type AuthInfo interface {
	GetHost() string
	GetPort() int32
	GetDbUsername() string
	GetPassword() string
	GetSslMode() string
}

type AuthInfoDetail struct {
	host       string
	port       int32
	dbUsername string
	password   string
	sslMode    string
}

func (a AuthInfoDetail) GetHost() string {
	return a.host
}

func (a AuthInfoDetail) GetPort() int32 {
	return a.port
}

func (a AuthInfoDetail) GetDbUsername() string {
	return a.dbUsername
}

func (a AuthInfoDetail) GetPassword() string {
	return a.password
}

func (a AuthInfoDetail) GetSslMode() string {
	return a.sslMode
}

func NewAuthInfo(host string, port int32, dbUsername string, password string, sslMode string) AuthInfoDetail {
	authInfo := AuthInfoDetail{
		host:       host,
		port:       port,
		dbUsername: dbUsername,
		password:   password,
		sslMode:    sslMode,
	}

	return authInfo
}
