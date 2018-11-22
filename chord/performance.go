package chord

import (
	"fmt"
	"math/rand"
	"time"
)

type CPUPerformance struct {
	/*
		CPU Utilization from a node will be stored in an object of this type
		Node (string): The identifier of the node creating the object
		TimeElapsed (time.Duration): CPU Utilization of the node
		Operation (string): The operation that the CPU Utilization corresponds to
		Configuration (Config): The details of the run (number of runs, number of events, order of events, etc.)
	*/
	Node          string
	TimeElapsed   time.Duration
	Operation     string
	Configuration Config
}

type QueryPerformance struct {
	/*
		Performance metric of batch queries is stored in an object of this type
		NumberOfQueries (int): The number of queries in the batch
		TimeElapsed (time.Duration): Time elapsed in processing the batch
	*/
	Run             int
	NumberOfQueries int
	TimeElapsed     time.Duration
}

var CPUPerformanceMetrics []CPUPerformance
var QueryPerformanceMetrics []QueryPerformance

func InitPerformance() {
	rand.Seed(time.Now().UnixNano())
}

func MetricsHandler(node string, timeElapsed time.Duration, operation string, config Config) {
	performance := CPUPerformance{
		Node:          node,
		TimeElapsed:   timeElapsed,
		Operation:     operation,
		Configuration: config,
	}
	CPUPerformanceMetrics = append(CPUPerformanceMetrics, performance)
}



func (ring *Ring) TestPerformance(n, nQ, qS, nQS int) {
	/*
		This function performs tests on randomly generated queries.
		Input:
			ring (*Ring): The ring on which the tests are to be performed
			num (int): Number of runs that need to be performed on the ring
		Output:
			This function fills global object queryPerformanceMetric with the metrics collected during testing
	*/
	numQueries := nQ
	for i := 0; i < nQS; i++ {
		for j := 0; j < n; j++ {
			queries := generateQueries(numQueries)
			start := time.Now()
			for _, query := range queries {
				if query[0] == "GET" {
					nodes, err := ring.Lookup(1, []byte(query[1]))
					if err != nil {
						fmt.Println("Error in Lookup:", err.Error())
					}
					localNode, err := ring.GetLocalNode(nodes[0])
					if err != nil {
						fmt.Println("Node not found")
					}
					value, err := localNode.DataStore.Get(query[1])
					if err != nil {
						fmt.Println(err.Error())
					} else {
						fmt.Println(string(value))
					}
				} else if query[0] == "SET" {
					nodes, err := ring.Lookup(1, []byte(query[1]))
					if err != nil {
						fmt.Println("Error in Lookup:", err.Error())
					}
					localNode, err := ring.GetLocalNode(nodes[0])
					if err != nil {
						fmt.Println("Node not found")
					}
					err = localNode.DataStore.Set(query[1], query[2])
					if err != nil {
						fmt.Println(err.Error())
					} else {
						fmt.Println("True")
					}
				} else if query[0] == " DELETE" {
					nodes, err := ring.Lookup(1, []byte(query[1]))
					if err != nil {
						fmt.Println("Error in Lookup:", err.Error())
					}
					localNode, err := ring.GetLocalNode(nodes[0])
					if err != nil {
						fmt.Println("Node not found")
					}
					err = localNode.DataStore.Delete(query[1])
					if err != nil {
						fmt.Println(err.Error())
					} else {
						fmt.Println("True")
					}
				}
			}
			queryPerformanceMetric := QueryPerformance{
				Run:             j,
				NumberOfQueries: numQueries,
				TimeElapsed:     time.Since(start),
			}
			QueryPerformanceMetrics = append(QueryPerformanceMetrics, queryPerformanceMetric)
		}
		numQueries += qS
	}
}

func generateQueries(num int) [][]string {
	/*
		This function generates a random sequence of queries (GET, SET, DELETE)
		Input:
			num (int): the number of queries needed to be generated
		Output:
			queries ([][]string): Returns an array of Queries of the form (["Key", "Value"])
	*/
	var queries [][]string
	possibleTypesOfQueries := []string{"GET", "SET", "DELETE"}
	for i := 0; i < num; i++ {
		val := rand.Intn(len(possibleTypesOfQueries))
		var query []string
		query = append(query, possibleTypesOfQueries[val])
		if query[0] == "GET" || query[0] == "DELETE" {
			query = append(query, RandStringRunes(8))
		} else {
			query = append(query, RandStringRunes(8))
			query = append(query, RandStringRunes(8))
		}
		queries = append(queries, query)
	}
	return queries
}

func RandStringRunes(n int) string {
	/*
		Generates a random string of required length
		Input:
			n (int): Length of the string required
		Output:
			b (string): A random string of length n
	*/
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (r *Ring) Simulate(n int) {
	var keys []string
	for i := 0; i < n; i++ {
		key := RandStringRunes(8)
		keys = append(keys, key)
	}
	for _, key := range keys {
		_, err := r.Lookup(1, []byte(key))
		if err != nil {
			fmt.Println("Error in Lookup:", err.Error())
		}
		// To implement get trace
	}
}
