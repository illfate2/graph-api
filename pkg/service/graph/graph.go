package graph

import (
	"sort"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/graph/simple"

	"github.com/illfate2/graph-api/pkg/service/graph/paths"

	"github.com/illfate2/graph-api/pkg/model"
)

type (
	IncidenceMatrix map[model.Edge]map[model.Node]int
	AdjacencyMatrix map[model.Node]map[model.Node]int
)

type Methods interface {
	IncidenceMatrix(graph model.Graph) IncidenceMatrix
	PlanarCheck(graph model.Graph) bool
	findEccentricity(graph model.Graph, node model.Node, nodes map[model.Node]struct{}) uint64
	FindDiameter(graph model.Graph) uint64
	FindRadius(graph model.Graph) uint64
	FindCenter(graph model.Graph) []model.Node
	AdjacencyMatrix(graph model.Graph) AdjacencyMatrix
	ShortestPath(graph model.Graph, fromNode, toNode uint64) []model.Node
	AllShortestPaths(graph model.Graph, fromNode, toNode uint64) [][]model.Node
	HamiltonianPath(graph model.Graph, orig uint64) ([]model.Node, bool)
	EulerianCycle(graph model.Graph, orig uint64) ([]model.Node, bool)
	Cartesian(first, second model.Graph) model.Graph
}

type Graph struct {
}

func (g Graph) PlanarCheck(graph model.Graph) bool {
	if len(graph.Edges) <= (3*len(setNodes(graph)) - 6) {
		return true
	}
	return false
}

func (g Graph) IncidenceMatrix(graph model.Graph) IncidenceMatrix {
	nodes := setNodes(graph)
	edges := make(IncidenceMatrix)
	for _, e := range graph.Edges {
		mNodes := make(map[model.Node]int)
		for n := range nodes {
			if e.IsDirected {
				mNodes[n] = 0
				if e.From == n {
					mNodes[n] = 1
				}
				if e.To == n {
					if mNodes[n] == 0 {
						mNodes[n] = -1
					}
				}
			} else {
				mNodes[n] = 0
				if e.From == n {
					mNodes[n] = 1
				}
				if e.To == n {
					mNodes[n] = 1
				}
			}
		}
		edges[e] = mNodes
	}
	return edges
}

func (g Graph) AdjacencyMatrix(graph model.Graph) AdjacencyMatrix {
	nodes := make(map[model.Node][]model.Node)
	for _, e := range graph.Edges {
		nodes[e.From] = []model.Node{}
		nodes[e.To] = []model.Node{}
	}
	for _, e := range graph.Edges {
		if e.IsDirected {
			nodes[e.From] = append(nodes[e.From], e.To)
		} else {
			nodes[e.From] = append(nodes[e.From], e.To)
			nodes[e.To] = append(nodes[e.To], e.From)
		}
	}
	matrix := make(AdjacencyMatrix)
	for n, s := range nodes {
		nodeCount := make(map[model.Node]int)
		for n2 := range nodes {
			nodeCount[n2] = 0
			for _, nInS := range s {
				if nInS == n2 {
					nodeCount[n2] = 1
					break
				}
			}
		}
		matrix[n] = nodeCount
	}
	return matrix
}

func (g Graph) ShortestPath(graph model.Graph, fromNode, toNode uint64) []model.Node {
	matrix := g.AdjacencyMatrix(graph)
	nodes := setNodes(graph)
	var currentNode model.Node
	for node := range nodes {
		if node.ID == fromNode {
			currentNode = node
			break
		}
	}

	shortestPaths := paths.AllShortestPathsFind(nodes, matrix, currentNode, toNode)
	if shortestPaths == nil {
		return nil
	}
	return shortestPaths[0]
}

// Returns eccentricity of node or
// 0 if graph is disconnected
func (g Graph) findEccentricity(graph model.Graph, node model.Node, nodes map[model.Node]struct{}) uint64 {
	var maxCost uint64 = 0
	for toNode := range nodes {
		if node != toNode {
			path := g.ShortestPath(graph, node.ID, toNode.ID)

			if path == nil {
				return 0
			}
			cost := uint64(len(path) - 1)

			if cost > maxCost {
				maxCost = cost
			}
		}
	}
	return maxCost
}

// Returns diameter of the graph or
// 0 if graph is disconnected
func (g Graph) FindDiameter(graph model.Graph) uint64 {
	nodes := setNodes(graph)
	var maxCost uint64 = 0

	for fromNode := range nodes {
		eccentricity := g.findEccentricity(graph, fromNode, nodes)
		if eccentricity == 0 {
			return 0
		}
		if eccentricity > maxCost {
			maxCost = eccentricity
		}
	}
	return maxCost
}

// Returns radius of the graph or
// 0 if graph is disconnected
func (g Graph) FindRadius(graph model.Graph) uint64 {
	nodes := setNodes(graph)
	minCost := uint64(len(nodes) + 2)

	for fromNode := range nodes {
		eccentricity := g.findEccentricity(graph, fromNode, nodes)
		if eccentricity == 0 {
			return 0
		}

		if eccentricity < minCost {
			minCost = eccentricity
		}
	}
	return minCost
}

// Returns center of the graph or
// nil if graph is disconnected
func (g Graph) FindCenter(graph model.Graph) []model.Node {
	nodes := setNodes(graph)
	minCost := uint64(len(nodes) + 2)
	var centerNodes []model.Node
	eccentricities := make(map[model.Node]uint64)

	for fromNode := range nodes {
		eccentricities[fromNode] = g.findEccentricity(graph, fromNode, nodes)
		if eccentricities[fromNode] == 0 {
			return nil
		}
		if eccentricities[fromNode] <= minCost {
			minCost = eccentricities[fromNode]
		}
	}

	for currentNode := range eccentricities {
		if eccentricities[currentNode] == minCost {
			centerNodes = append(centerNodes, currentNode)
		}
	}

	return centerNodes
}

func (g Graph) AllShortestPaths(graph model.Graph, fromNode, toNode uint64) [][]model.Node {
	matrix := g.AdjacencyMatrix(graph)
	nodes := setNodes(graph)
	var currentNode model.Node
	for node := range nodes {
		if node.ID == fromNode {
			currentNode = node
			break
		}
	}

	return paths.AllShortestPathsFind(nodes, matrix, currentNode, toNode)
}

func (g Graph) HamiltonianPath(graph model.Graph, orig uint64) ([]model.Node, bool) {
	visited := make(map[uint64]bool)
	path := []uint64{orig}
	nodeToEdges := make(map[uint64]map[uint64]struct{})
	for _, e := range graph.Edges {
		nodeToEdges[e.From.ID] = make(map[uint64]struct{})
		nodeToEdges[e.To.ID] = make(map[uint64]struct{})
	}
	for _, e := range graph.Edges {
		nodeToEdges[e.From.ID][e.To.ID] = struct{}{}
		nodeToEdges[e.To.ID][e.From.ID] = struct{}{}
	}

	hamiltonPath, find := g.hamiltonianPath(orig, orig, visited, path, nodeToEdges)
	if !find {
		return nil, false
	}
	nodes := graphToNodes(graph)
	resPath := make([]model.Node, 0, len(hamiltonPath))
	for _, n := range hamiltonPath {
		resNode := nodes[n]
		resPath = append(resPath, resNode)
	}
	return resPath, true
}

func (g Graph) EulerianCycle(graph model.Graph, startedNode uint64) ([]model.Node, bool) {
	nodeToEdges := make(map[uint64]map[uint64]struct{})
	for _, e := range graph.Edges {
		nodeToEdges[e.From.ID] = make(map[uint64]struct{})
		nodeToEdges[e.To.ID] = make(map[uint64]struct{})
	}
	for _, e := range graph.Edges {
		nodeToEdges[e.From.ID][e.To.ID] = struct{}{}
		nodeToEdges[e.To.ID][e.From.ID] = struct{}{}
	}

	unvisitedNodeToEdges := make(map[uint64]map[uint64]struct{})
	for _, e := range graph.Edges {
		unvisitedNodeToEdges[e.From.ID] = make(map[uint64]struct{})
		unvisitedNodeToEdges[e.To.ID] = make(map[uint64]struct{})
	}
	for _, e := range graph.Edges {
		unvisitedNodeToEdges[e.From.ID][e.To.ID] = struct{}{}
		unvisitedNodeToEdges[e.To.ID][e.From.ID] = struct{}{}
	}
	for _, e := range nodeToEdges {
		if len(e)%2 != 0 {
			return nil, false
		}
	}

	// Hierholzer's algorithm
	var currentVertex, nextVertex uint64

	tour := []uint64{}
	stack := []uint64{startedNode}

	for len(stack) > 0 {
		currentVertex = stack[len(stack)-1]
		// Get an arbitrary edge from the current vertex
		if len(unvisitedNodeToEdges[currentVertex]) > 0 {
			for nextVertex = range unvisitedNodeToEdges[currentVertex] {
				break
			}
			delete(unvisitedNodeToEdges[currentVertex], nextVertex)
			delete(unvisitedNodeToEdges[nextVertex], currentVertex)
			stack = append(stack, nextVertex)
		} else {
			tour = append(tour, stack[len(stack)-1])
			stack = stack[:len(stack)-1]
		}
	}

	nodes := graphToNodes(graph)
	resPath := make([]model.Node, 0, len(tour))
	for _, n := range tour {
		resNode := nodes[n]
		resPath = append(resPath, resNode)
	}
	return resPath, true
}

func (g Graph) hamiltonianPath(
	orig, dest uint64,
	visited map[uint64]bool,
	path []uint64,
	nodeToEdges map[uint64]map[uint64]struct{},
) ([]uint64, bool) {
	if len(visited) == len(nodeToEdges) {
		if path[len(path)-1] == dest {
			return path, true
		}

		return nil, false
	}

	for tv := range nodeToEdges[orig] {
		if _, ok := visited[tv]; !ok && (dest != tv || len(visited) == len(nodeToEdges)-1) {
			visited[tv] = true
			path = append(path, tv)
			if path, found := g.hamiltonianPath(tv, dest, visited, path, nodeToEdges); found {
				return path, true
			}
			path = path[:len(path)-1]
			delete(visited, tv)
		}
	}

	return nil, false
}

func (g Graph) Cartesian(first, second model.Graph) model.Graph {
	return model.Graph{}
}

func toUndirectedGraph(g model.Graph) *simple.UndirectedGraph {
	undirGraph := simple.NewUndirectedGraph()

	for _, e := range g.Edges {
		undirGraph.SetEdge(simple.Edge{
			F: simple.Node(e.From.ID),
			T: simple.Node(e.To.ID),
		})
	}
	return undirGraph
}

func (m AdjacencyMatrix) String() string {
	var strBuilder strings.Builder
	nodes := m.getSortedSlice()
	strBuilder.WriteString("   ")

	for _, n := range nodes {
		id := strconv.FormatUint(n.ID, 10)
		strBuilder.WriteString(id)
		strBuilder.WriteString(" ")
	}
	strBuilder.WriteString("\n")
	for _, n := range nodes {
		nodesToIndex := m[n]
		strBuilder.WriteString(strconv.FormatUint(n.ID, 10))
		strBuilder.WriteString(": ")

		for _, n := range nodes {
			idx := nodesToIndex[n]
			id := strconv.FormatInt(int64(idx), 10)
			strBuilder.WriteString(id)
			strBuilder.WriteString(" ")
		}
		strBuilder.WriteString("\n")
	}

	return strBuilder.String()

}

func setNodes(graph model.Graph) map[model.Node]struct{} {
	nodes := make(map[model.Node]struct{})
	for _, e := range graph.Edges {
		nodes[e.From] = struct{}{}
		nodes[e.To] = struct{}{}
	}
	return nodes
}

func graphToNodes(graph model.Graph) map[uint64]model.Node {
	nodes := make(map[uint64]model.Node)
	for _, e := range graph.Edges {
		nodes[e.From.ID] = e.From
		nodes[e.To.ID] = e.To
	}
	return nodes
}

func (m AdjacencyMatrix) getSortedSlice() []model.Node {
	nodes := make([]model.Node, 0, len(m))
	for k := range m {
		nodes = append(nodes, k)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})

	return nodes
}
