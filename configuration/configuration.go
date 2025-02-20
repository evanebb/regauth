package configuration

type Configuration struct {
	Log      Log
	HTTP     HTTP
	Database Database
	Token    Token
	Auth     Auth
	Registry Registry
}

type Log struct {
	Level     string
	Formatter string
}

type HTTP struct {
	Addr        string
	Certificate string
	Key         string
	SessionKey  string
}

type Database struct {
	Host     string
	Name     string
	User     string
	Password string
	Port     int
}

type Token struct {
	Issuer      string
	Service     string
	Certificate string
	Key         string
	Alg         string
}

type Auth struct {
	Local LocalAuth
}

type LocalAuth struct {
	enabled bool
}

type Registry struct {
	Host string
}
