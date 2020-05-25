package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/illfate2/graph-api/pkg/model"
	"github.com/illfate2/graph-api/pkg/service"
	"github.com/illfate2/graph-api/pkg/service/graph"
)

type Server struct {
	http.Handler
	r       *mux.Router
	service service.Service
}

func New(service service.Service) *Server {
	r := mux.NewRouter()
	s := &Server{
		r:       r,
		service: service,
	}
	r.HandleFunc("api/v1/graph", s.CreateGraph).Methods(http.MethodPost)
	r.HandleFunc("api/v1/graph/{id:[1-9]+[0-9]*}", s.Graph).Methods(http.MethodGet)
	r.HandleFunc("api/v1/graph/{id:[1-9]+[0-9]*}", s.UpdateGraph).Methods(http.MethodPut)
	r.HandleFunc("api/v1/graph/{id:[1-9]+[0-9]*}", s.DeleteGraph).Methods(http.MethodDelete)
	r.HandleFunc("api/v1/graph/{id:[1-9]+[0-9]*}/adjacencyMatrix", s.AdjacencyMatrix).Methods(http.MethodGet)
	r.HandleFunc("api/v1/graph/{id:[1-9]+[0-9]*}/incidenceMatrix", s.IncidenceMatrix).Methods(http.MethodGet)
	r.HandleFunc("api/v1/graph/{id:[1-9]+[0-9]*}", s.Graph).Methods(http.MethodGet)

	return s
}

func (s *Server) CreateGraph(w http.ResponseWriter, req *http.Request) {
	var g model.Graph
	err := json.NewDecoder(req.Body).Decode(&g)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	id, err := s.service.CreateGraph(g)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		ID uint64 `json:"id"`
	}{
		ID: id,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) Graph(w http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	g, err := s.service.Graph(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(g)
}

func (s *Server) UpdateGraph(w http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var g model.Graph
	err = json.NewDecoder(req.Body).Decode(&g)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	g.ID = id

	err = s.service.UpdateGraph(g)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) DeleteGraph(w http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.service.DeleteGraph(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) IncidenceMatrix(w http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	m, err := s.service.IncidenceMatrix(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		Matrix graph.IncidenceMatrix `json:"matrix"`
	}{
		Matrix: m,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) AdjacencyMatrix(w http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	m, err := s.service.AdjacencyMatrix(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		Matrix graph.AdjacencyMatrix `json:"matrix"`
	}{
		Matrix: m,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func getID(req *http.Request) (uint64, error) {
	vars := mux.Vars(req)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	return id, err
}
