package rpc

//go:generate zenrpc

import (
	"github.com/semrush/zenrpc"
	"github.com/yuuki0xff/clustertest/databases"
	"github.com/yuuki0xff/clustertest/models"
	"net/http"
)

type Server struct {
	DB models.TaskDB
} //zenrpc

func ServeServer(listenAddr string, db models.TaskDB) error {
	srv := &Server{
		DB: db,
	}
	s := zenrpc.NewServer(zenrpc.Options{})
	s.Register("", srv)
	return http.ListenAndServe(listenAddr, s)
}

func (s *Server) Run_Task(spec []byte) string {
	id, err := s.DB.Create(&databases.MemTask{
		Spec: spec,
	})
	if err != nil {
		panic(err)
	}
	return id.String()
}
func (s *Server) Is_Ready_Task(id string) bool {
	tid := &databases.StringTaskID{
		ID: id,
	}
	detail, err := s.DB.Inspect(tid)
	if err != nil {
		panic(err)
	}
	return detail.State() == "finished"
}
func (s *Server) Get_Task_Result(id string) models.TaskResult {
	tid := &databases.StringTaskID{
		ID: id,
	}
	detail, err := s.DB.Inspect(tid)
	if err != nil {
		panic(err)
	}
	return detail.Result()
}
