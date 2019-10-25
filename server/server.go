package server

type Center struct {
	Server []*Server
}

type Server struct {
	Host string
	Port int
	Env  string
}

func (s *Server) Register() {

}
