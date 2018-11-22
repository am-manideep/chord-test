package chord

import (
	"fmt"
	"time"
)

func TestCorrectness(n, minST, maxST, sTS, nSTS, eFD, eFDS, nEFDS int) (bool, error) {
	finalPass := true
	config := DefaultConfig("local")
	trans := InitLocalTransport(nil)
	stabilizeMin := time.Duration(minST) * time.Second
	stabilizeMax := time.Duration(maxST) * time.Second
	for i := 0; i < nSTS; i++ {
		sleep := time.Duration(eFD) * time.Second
		for j := 0; j < nEFDS; j++ {
			config.StabilizeMax = stabilizeMax
			config.StabilizeMin = stabilizeMin

			ring, err := Create(config, trans)
			if err != nil {
				fmt.Println("error in creating ring:", err.Error())
				return false, err
			}
			for k:=0; k<n; k++ {
				pass := ring.CheckCorrectness(100, sleep)
				if !pass {
					finalPass = pass
				}
			}

			sleep += time.Duration(eFDS) * time.Second
		}
		stabilizeMax += time.Duration(sTS) * time.Second
		stabilizeMin += time.Duration(sTS) * time.Second
	}
	return finalPass, nil
}

func CheckCorrectnessInvariants(ring *Ring) bool {
	pass := true
	ok := connectedAppendages(ring)
	if !ok {
		pass = false
		fmt.Println("Connected Appendages Invariant failed")
	}
	ok = atleastOneRing(ring)
	if !ok {
		pass = false
		fmt.Println("Atleast One Ring Invariant failed")
	}
	ok = orderedRing(ring)
	if !ok {
		pass = false
		fmt.Println("Ordered Ring Invariant failed")
	}
	ok = atmostOneRing(ring)
	if !ok {
		pass = false
		fmt.Println("Atmost One Ring Invariant failed")
	}
	return pass
}

func CheckConsistencyInvariants(ring *Ring) bool {
	pass := true
	ok := orderedMerges(ring)
	if !ok {
		pass = false
		fmt.Println("Connected Appendages Invariant failed")
	}
	ok = orderedAppendages(ring)
	if !ok {
		pass = false
		fmt.Println("Connected Appendages Invariant failed")
	}
	ok = validSuccessorList(ring)
	if !ok {
		pass = false
		fmt.Println("Connected Appendages Invariant failed")
	}
	return pass
}

//Correctness Invariants
func connectedAppendages(ring *Ring) bool {
	/*
		An appendage is a node that is connected to the ring externally, that is by only one other node.
		Addition of a new node results in an appendage.
		This invariant asserts whether all the appendages are connected.
	*/
	return true
}

func atleastOneRing(ring *Ring) bool {
	/*
		This invariant asserts that the ring has at least one node, or in other words the ring exists.
	*/
	return true
}

func orderedRing(ring *Ring) bool {
	/*
		This invariant asserts whether the ring is always ordered by identifiers.
	*/
	return true
}

func atmostOneRing(ring *Ring) bool {
	/*
		This invariant asserts that only one ring is formed in the system.
	*/
	return true
}

//Key Data Consistency Invariants
func orderedMerges(ring *Ring) bool {
	/*
		This invariant asserts whether the appendages are merged into the ring at the right places
	*/
	return true
}

func orderedAppendages(ring *Ring) bool {
	/*
		This invariant asserts whether the order imposed by successors is consistent with identifier order, in the cycle and in the appendages respectively
	*/
	return true
}

func validSuccessorList(ring *Ring) bool {
	/*
		This invariant asserts that if v and w are members, and if w’s successor list skips over v, then v is not in the successor list of any immediate antecedent of w
	*/
	return true
}
