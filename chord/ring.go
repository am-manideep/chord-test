package chord

import (
	"bytes"
	"correct-chord-go/global"
	"crypto/sha1"
	"errors"
	"fmt"
	"hash"
	"log"
	"math/rand"
	"sort"
	"time"
)

// Configuration for Chord nodes
type Config struct {
	Hostname      string           // Local host name
	NumVnodes     int              // Number of vnodes per physical chord
	HashFunc      func() hash.Hash // Hash function to use
	StabilizeMin  time.Duration    // Minimum stabilization time
	StabilizeMax  time.Duration    // Maximum stabilization time
	NumSuccessors int              // Number of successors to maintain
	Delegate      Delegate         // Invoked to handle ring events
	hashBits      int              // Bit size of the hash function
}

// Stores the state required for a Chord ring
type Ring struct {
	config     *Config
	transport  Transport
	vnodes     []*localVnode
	delegateCh chan func()
	shutdown   chan bool
}

func (r *Ring) init(conf *Config, trans Transport) {
	// Set our variables
	r.config = conf
	r.vnodes = make([]*localVnode, conf.NumVnodes)
	r.transport = InitLocalTransport(trans)
	r.delegateCh = make(chan func(), 32)

	// Initializes the vnodes
	for i := 0; i < conf.NumVnodes; i++ {
		vn := &localVnode{}
		r.vnodes[i] = vn
		vn.ring = r
		vn.init(i)
	}

	// Sort the vnodes
	sort.Sort(r)
}

// Len is the number of vnodes
func (r *Ring) Len() int {
	return len(r.vnodes)
}

// Less returns whether the vnode with index i should sort
// before the vnode with index j.
func (r *Ring) Less(i, j int) bool {
	return bytes.Compare(r.vnodes[i].Id, r.vnodes[j].Id) == -1
}

// Swap swaps the vnodes with indexes i and j.
func (r *Ring) Swap(i, j int) {
	r.vnodes[i], r.vnodes[j] = r.vnodes[j], r.vnodes[i]
}

// Returns the nearest local vnode to the key
func (r *Ring) nearestVnode(key []byte) *localVnode {
	for i := len(r.vnodes) - 1; i >= 0; i-- {
		if bytes.Compare(r.vnodes[i].Id, key) == -1 {
			return r.vnodes[i]
		}
	}
	// Return the last vnode
	return r.vnodes[len(r.vnodes)-1]
}

// Schedules each vnode in the ring
func (r *Ring) schedule() {
	if r.config.Delegate != nil {
		go r.delegateHandler()
	}
	for i := 0; i < len(r.vnodes); i++ {
		r.vnodes[i].schedule()
	}
}

// Wait for all the vnodes to shutdown
func (r *Ring) stopVnodes() {
	r.shutdown = make(chan bool, r.config.NumVnodes)
	for i := 0; i < r.config.NumVnodes; i++ {
		<-r.shutdown
	}
}

// Stops the delegate handler
func (r *Ring) stopDelegate() {
	if r.config.Delegate != nil {
		// Wait for all delegate messages to be processed
		<-r.invokeDelegate(r.config.Delegate.Shutdown)
		close(r.delegateCh)
	}
}

// Initializes the vnodes with their local successors
func (r *Ring) setLocalSuccessors() {
	numV := len(r.vnodes)
	numSuc := global.Min(r.config.NumSuccessors, numV-1)
	for idx, vnode := range r.vnodes {
		for i := 0; i < numSuc; i++ {
			vnode.successors[i] = &r.vnodes[(idx+i+1)%numV].Vnode
		}
	}
}

// Invokes a function on the delegate and returns completion channel
func (r *Ring) invokeDelegate(f func()) chan struct{} {
	if r.config.Delegate == nil {
		return nil
	}

	ch := make(chan struct{}, 1)
	wrapper := func() {
		defer func() {
			ch <- struct{}{}
		}()
		f()
	}

	r.delegateCh <- wrapper
	return ch
}

// This handler runs in a go routine to invoke methods on the delegate
func (r *Ring) delegateHandler() {
	for {
		f, ok := <-r.delegateCh
		if !ok {
			break
		}
		r.safeInvoke(f)
	}
}

// Called to safely call a function on the delegate
func (r *Ring) safeInvoke(f func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Caught a panic invoking a delegate function! Got: %s", r)
		}
	}()
	f()
}

// Returns the default Ring configuration
func DefaultConfig(hostname string) *Config {
	return &Config{
		hostname,
		8,        // 8 vnodes
		sha1.New, // SHA1
		time.Duration(15 * time.Second),
		time.Duration(45 * time.Second),
		2,   // 8 successors
		nil, // No delegate
		160, // 160bit hash function
	}
}

// Creates a new Chord ring given the config and transport
func Create(conf *Config, trans Transport) (*Ring, error) {
	// Initialize the hash bits
	conf.hashBits = conf.HashFunc().Size() * 8

	// Create and initialize a ring
	ring := &Ring{}
	ring.init(conf, trans)
	ring.setLocalSuccessors()
	ring.schedule()
	fmt.Println(ring.vnodes)
	for _, node := range ring.vnodes {
		fmt.Println((*node).successors)
	}
	return ring, nil
}

// Joins an existing Chord ring
func Join(conf *Config, trans Transport, existing string) (*Ring, error) {
	// Initialize the hash bits
	conf.hashBits = conf.HashFunc().Size() * 8

	// Request a list of Vnodes from the remote host
	hosts, err := trans.ListVnodes(existing)
	if err != nil {
		return nil, err
	}
	if hosts == nil || len(hosts) == 0 {
		return nil, fmt.Errorf("Remote host has no vnodes!")
	}

	// Create a ring
	ring := &Ring{}
	ring.init(conf, trans)

	// Acquire a live successor for each Vnode
	for _, vn := range ring.vnodes {
		// Get the nearest remote vnode
		nearest := NearestVnodeToKey(hosts, vn.Id)

		// Query for a list of successors to this Vnode
		succs, err := trans.FindSuccessors(nearest, conf.NumSuccessors, vn.Id)
		if err != nil {
			return nil, fmt.Errorf("Failed to find successor for vnodes! Got %s", err)
		}
		if succs == nil || len(succs) == 0 {
			return nil, fmt.Errorf("Failed to find successor for vnodes! Got no vnodes!")
		}

		// Assign the successors
		for idx, s := range succs {
			vn.successors[idx] = s
		}
	}

	// Start delegate handler
	if ring.config.Delegate != nil {
		go ring.delegateHandler()
	}

	// Do a fast stabilization, will schedule regular execution
	for _, vn := range ring.vnodes {
		vn.stabilize()
	}
	return ring, nil
}

// Leaves a given Chord ring and shuts down the local vnodes
func (r *Ring) Leave() error {
	// Shutdown the vnodes first to avoid further stabilization runs
	r.stopVnodes()

	// Instruct each vnode to leave
	var err error
	for _, vn := range r.vnodes {
		err = global.MergeErrors(err, vn.leave())
	}

	// Wait for the delegate callbacks to complete
	r.stopDelegate()
	return err
}

// Shutdown shuts down the local processes in a given Chord ring
// Blocks until all the vnodes terminate.
func (r *Ring) Shutdown() {
	r.stopVnodes()
	r.stopDelegate()
}

// Does a key lookup for up to N successors of a key
func (r *Ring) Lookup(n int, key []byte) ([]*Vnode, error) {
	// Ensure that n is sane
	if n > r.config.NumSuccessors {
		return nil, fmt.Errorf("Cannot ask for more successors than NumSuccessors!")
	}

	// Hash the key
	h := r.config.HashFunc()
	h.Write(key)
	key_hash := h.Sum(nil)

	// Find the nearest local vnode
	nearest := r.nearestVnode(key_hash)

	// Use the nearest chord for the lookup
	successors, err := nearest.FindSuccessors(n, key_hash)
	if err != nil {
		return nil, err
	}

	// Trim the nil successors
	for successors[len(successors)-1] == nil {
		successors = successors[:len(successors)-1]
	}
	return successors, nil
}

// Generates a random stabilization time
func RandStabilize(conf *Config) time.Duration {
	min := conf.StabilizeMin
	max := conf.StabilizeMax
	r := rand.Float64()
	return time.Duration((r * float64(max-min)) + float64(min))
}

// Returns the vnode nearest a key
func NearestVnodeToKey(vnodes []*Vnode, key []byte) *Vnode {
	for i := len(vnodes) - 1; i >= 0; i-- {
		if bytes.Compare(vnodes[i].Id, key) == -1 {
			return vnodes[i]
		}
	}
	// Return the last vnode
	return vnodes[len(vnodes)-1]
}

func (r *Ring) GetLocalNode(vnode *Vnode) (*localVnode, error) {
	for _, node := range r.vnodes {
		if bytes.Equal(node.Id, vnode.Id) {
			return node, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *Ring) PrintData() {
	for _, node := range r.vnodes {
		fmt.Println(node.DataStore)
	}
}

func (r *Ring) CheckCorrectness(num int, sleep time.Duration) bool {
	done := make(chan bool)
	go r.checkCorrectness(num, sleep, done)
	pass := <-done
	return pass
}

func (r *Ring) checkCorrectness(num int, sleep time.Duration, done chan bool) {
	/*
		This function checks correctness by asserting the variants defined in correctness.go
		Input:
			r (*Ring): the ring whose correctness is to be tested
		Output:
			The function will make assertions about Correctness invariants and send the result to a log
	*/
	events := r.generateEvents(num)
	val := rand.Intn(len(r.vnodes))
	for _, event := range events {
		if event == "join" {
			r.vnodes[val].join()
		}
		if event == "leave" {
			r.vnodes[val].leave()
		}
		if event == "join" {
			r.vnodes[val].fail()
		}
		CheckConsistencyInvariants(r)
		time.Sleep(sleep)
	}
	pass := CheckCorrectnessInvariants(r)
	done <- pass
}

func (r *Ring) generateEvents(num int) []string {
	/*
		Generates a random sequence of events
		Input:
			num (int): number of events required
		Output:
			events ([]string): a sequence of randomly generated events, of length num
	*/
	possibleEvents := []string{"join", "leave", "fail"}
	events := []string{}
	for i := 0; i < num; i++ {
		val := rand.Intn(len(possibleEvents))
		events = append(events, possibleEvents[val])
	}
	return events
}
