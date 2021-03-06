package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/illfate2/graph-api/pkg/model"
	"github.com/illfate2/graph-api/pkg/service"
)

type Server struct {
	http.Handler
	service service.Service
}

func New(service service.Service) *Server {
	r := mux.NewRouter()
	s := Server{
		service: service,
		Handler: r,
	}

	r.Methods(http.MethodOptions).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {})
	r.Use(CORS)

	r.HandleFunc("/api/v1/graph", s.CreateGraph).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}", s.Graph).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}", s.UpdateGraph).Methods(http.MethodPut)
	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}", s.DeleteGraph).Methods(http.MethodDelete)
	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/adjacencyMatrix", s.AdjacencyMatrix).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/incidenceMatrix", s.IncidenceMatrix).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/diameter", s.FindDiameter).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/radius", s.FindRadius).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/tree", s.Tree).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/center", s.FindCenter).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/graph/{ids:[1-9]+[0-9]*[,][1-9]+[0-9]*}/cartesian", s.Cartesian).Methods(http.MethodGet)

	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/shortestPath", s.ShortestPath).
		Queries("fromNode", "{fromNode}", "toNode", "{toNode}").Methods(http.MethodGet)

	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/allShortestPath", s.AllShortestPaths).
		Queries("fromNode", "{fromNode}", "toNode", "{toNode}").Methods(http.MethodGet)

	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/allPath", s.AllPaths).
		Queries("fromNode", "{fromNode}", "toNode", "{toNode}").Methods(http.MethodGet)

	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/hamiltonianPath", s.HamiltonianPath).
		Queries("startNode", "{startNode}").Methods(http.MethodGet)
	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/eulerianCycle", s.EulerianCycle).
		Queries("startNode", "{startNode}").Methods(http.MethodGet)

	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/planarCheck", s.PlanarCheck).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/planarReduction", s.PlanarReduction).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/graph/{id:[1-9]+[0-9]*}/isTree", s.IsTree).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/graph/list", s.GraphList).Methods(http.MethodGet)
	return &s
}

func (s *Server) CreateGraph(w http.ResponseWriter, req *http.Request) {
	var g model.Graph
	err := json.NewDecoder(req.Body).Decode(&g)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print("Error when decoding JSON: ", err)
		return
	}
	id, err := s.service.CreateGraph(g)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print("Error when creating graph: ", err)
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

func (s *Server) GraphList(w http.ResponseWriter, req *http.Request) {
	list, err := s.service.List()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(list)
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
		Matrix string `json:"matrix"`
	}{
		Matrix: m.String(),
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
		Matrix string `json:"matrix"`
	}{
		Matrix: m.String(),
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) ShortestPath(w http.ResponseWriter, req *http.Request) {
	args, err := getShortestPathArgs(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path, err := s.service.ShortestPath(args.graphID, args.fromNode, args.toNode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		Path []model.Node `json:"path"`
	}{
		Path: path,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) FindDiameter(w http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	d, err := s.service.FindDiameter(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		Diameter uint64 `json:"diameter"`
	}{
		Diameter: d,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) Cartesian(w http.ResponseWriter, req *http.Request) {
	firstID, secondID, err := getIDs(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c, err := s.service.Cartesian(firstID, secondID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		Cartesian model.Graph `json:"cartesian"`
	}{
		Cartesian: c,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) FindCenter(w http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c, err := s.service.FindCenter(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		Diameter []model.Node `json:"center"`
	}{
		Diameter: c,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) FindRadius(w http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r, err := s.service.FindRadius(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		Diameter uint64 `json:"radius"`
	}{
		Diameter: r,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) PlanarCheck(w http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := s.service.PlanarCheck(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		IsPlanar bool `json:"isPlanar"`
	}{
		IsPlanar: res,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) PlanarReduction(w http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := s.service.PlanarReduction(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		PlanarGraph model.Graph `json:"planarGraph"`
	}{
		PlanarGraph: res,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) Tree(w http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := s.service.Tree(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		Tree model.Graph `json:"tree"`
	}{
		Tree: res,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) IsTree(w http.ResponseWriter, req *http.Request) {
	id, err := getID(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res := s.service.IsTree(id)
	resp := struct {
		IsTree bool `json:"isTree"`
	}{
		IsTree: res,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) HamiltonianPath(w http.ResponseWriter, req *http.Request) {
	s.path(w, req, s.service.HamiltonianPath)
}

func (s *Server) EulerianCycle(w http.ResponseWriter, req *http.Request) {
	s.path(w, req, s.service.EulerianCycle)
}

type pathF func(graphID, startedNode uint64) ([]model.Node, error)

func (s *Server) path(w http.ResponseWriter, req *http.Request, f pathF) {
	args, err := getPathArgs(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path, err := f(args.graphID, args.startedNode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		Path []model.Node `json:"path"`
	}{
		Path: path,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) AllShortestPaths(w http.ResponseWriter, req *http.Request) {
	args, err := getShortestPathArgs(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path, err := s.service.AllShortestPaths(args.graphID, args.fromNode, args.toNode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		Path [][]model.Node `json:"paths"`
	}{
		Path: path,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) AllPaths(w http.ResponseWriter, req *http.Request) {
	args, err := getShortestPathArgs(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path, err := s.service.AllPaths(args.graphID, args.fromNode, args.toNode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := struct {
		Path [][]model.Node `json:"paths"`
	}{
		Path: path,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func getID(req *http.Request) (uint64, error) {
	return getSpecificID(req, "id")
}

func getIDs(req *http.Request) (first, second uint64, err error) {
	vars := mux.Vars(req)
	ids := strings.Split(vars["ids"], ",")
	first, err = strconv.ParseUint(ids[0], 10, 64)
	if err != nil {
		return
	}
	second, err = strconv.ParseUint(ids[1], 10, 64)
	return
}

type shortestPathArgs struct {
	graphID  uint64
	fromNode uint64
	toNode   uint64
}

func getShortestPathArgs(req *http.Request) (shortestPathArgs, error) {
	id, err := getID(req)
	if err != nil {
		return shortestPathArgs{}, err
	}
	fromNode, err := getSpecificID(req, "fromNode")
	if err != nil {
		return shortestPathArgs{}, err
	}
	toNode, err := getSpecificID(req, "toNode")
	if err != nil {
		return shortestPathArgs{}, err
	}
	return shortestPathArgs{
		graphID:  id,
		fromNode: fromNode,
		toNode:   toNode,
	}, nil
}

type pathArgs struct {
	graphID     uint64
	startedNode uint64
}

func getPathArgs(req *http.Request) (pathArgs, error) {
	id, err := getID(req)
	if err != nil {
		return pathArgs{}, err
	}
	startedNode, err := getSpecificID(req, "startNode")
	if err != nil {
		return pathArgs{}, err
	}
	return pathArgs{
		graphID:     id,
		startedNode: startedNode,
	}, nil
}

func getSpecificID(req *http.Request, idName string) (uint64, error) {
	vars := mux.Vars(req)
	id, err := strconv.ParseUint(vars[idName], 10, 64)
	return id, err
}
