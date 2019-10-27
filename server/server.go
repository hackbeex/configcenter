package server

type EnvType string

const (
	EnvDev  EnvType = "develop"
	EnvProd EnvType = "product"
	EnvTest EnvType = "test"
)

type Server struct {
	Id   string
	Host string
	Port int
	Env  EnvType
}
