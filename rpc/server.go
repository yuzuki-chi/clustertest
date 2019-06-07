package rpc

//go:generate zenrpc

import (
	"github.com/semrush/zenrpc"
	"net/http"
)

type Server struct{} //zenrpc

func ServeServer(listenAddr string) error {
	s := zenrpc.NewServer(zenrpc.Options{})
	s.Register("", &Server{})
	return http.ListenAndServe(listenAddr, s)
}

func (s *Server) Run_Task(spec []byte) string {
	// TODO
	return "OK"
}
func (s *Server) Is_Ready_Task(id string) bool {
	// TODO
	return true
}
func (s *Server) Get_Task_Result(id string) *Result {
	// TODO
	return nil
}
