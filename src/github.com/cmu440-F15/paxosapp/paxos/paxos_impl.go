package paxos

import (
	"encoding/json"
	"errors"
	"github.com/cmu440-F15/paxosapp/rpc/paxosrpc"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type paxosNode struct {
	store       map[string]interface{} //main data store
	highestSeen map[string]int         //highest prop # seen for each key
	va          map[string]interface{} //va value for prepare phase
	na          map[string]int         //na value for prepare phase

	nodeID     int
	hostMap    map[int]string
	myHostPort string
	allNodes   []*rpc.Client
	numNodes   int
}

type Na_va struct {
	na int
	va interface{}
}

// NewPaxosNode creates a new PaxosNode. This function should return only when
// all nodes have joined the ring, and should return a non-nil error if the node
// could not be started in spite of dialing the other nodes numRetries times.
//
// hostMap is a map from node IDs to their hostports, numNodes is the number
// of nodes in the ring, replace is a flag which indicates whether this node
// is a replacement for a node which failed.
func NewPaxosNode(myHostPort string, hostMap map[int]string, numNodes, srvId, numRetries int, replace bool) (PaxosNode, error) {
	hostMap[srvId] = myHostPort
	newNode := paxosNode{
		store:       make(map[string]interface{}),
		highestSeen: make(map[string]int),
		na:          make(map[string]int),
		va:          make(map[string]interface{}),
		nodeID:      srvId,
		hostMap:     hostMap,
		myHostPort:  myHostPort,
		allNodes:    make([]*rpc.Client, numNodes),
		numNodes:    numNodes}

	// Set up rpc calls
	rpc.RegisterName("PaxosNode", &newNode)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", myHostPort)
	if e != nil {
		return nil, errors.New("PaxosNode couldn't start listening")
	}
	go http.Serve(l, nil)

	// Connect to all other nodes
	for i := 0; i < numNodes; i++ {
		for j := 0; j < 5; j++ {
			client, err := rpc.DialHTTP("tcp", hostMap[i])
			if err != nil {
				if j <= 3 {
					time.Sleep(time.Second)
				} else {
					return nil, errors.New("PaxosNode couldn't connect to other nodes")
				}
			} else {
				// Add this rpc connection for this node
				newNode.allNodes[i] = client
				if replace && i != srvId {
					var rreply paxosrpc.ReplaceServerReply
					rargs := &paxosrpc.ReplaceServerArgs{SrvID: srvId, Hostport: myHostPort}
					client.Call("PaxosNode.RecvReplaceServer", rargs, &rreply)
				}
				break
			}

		}

	}
	//if were a replacement node, let's get caught up with the store.
	if replace {
		i := 0
		//don't ask ourselves, ask another node!
		if srvId == 0 {
			i = 1
		}
		client := newNode.allNodes[i]
		var catchupData map[string]interface{}
		var creply paxosrpc.ReplaceCatchupReply
		cargs := &paxosrpc.ReplaceCatchupArgs{}
		client.Call("PaxosNode.RecvReplaceCatchup", cargs, &creply)
		json.Unmarshal(creply.Data, &catchupData)
		newNode.store = catchupData

	}
	return &newNode, nil
}

func (pn *paxosNode) GetNextProposalNumber(args *paxosrpc.ProposalNumberArgs, reply *paxosrpc.ProposalNumberReply) error {
	key := args.Key
	var n int
	if m, found := pn.highestSeen[key]; found {
		n = (m/pn.numNodes+1)*pn.numNodes + pn.nodeID
	} else {
		n = pn.nodeID
	}
	pn.highestSeen[key] = n
	reply.N = n
	return nil
}

func (pn *paxosNode) Propose(args *paxosrpc.ProposeArgs, reply *paxosrpc.ProposeReply) error {
	doneChan := make(chan interface{})
	// Everything in a goroutine for timeout checking
	go func() {
		// Setup for prepare phase
		pArgs := &paxosrpc.PrepareArgs{Key: args.Key, N: args.N}
		var client *rpc.Client
		acceptChan := make(chan Na_va, pn.numNodes)

		// Send out prepare messages
		for i := 0; i < pn.numNodes; i++ {
			client = pn.allNodes[i]
			go sendProposal(pn, client, pArgs, acceptChan)
		}

		// Receive prepare messages
		highestna := 0
		var highestva interface{}
		for j := 0; j <= pn.numNodes/2; j++ {
			nava := <-acceptChan
			if nava.na >= highestna {
				highestna = nava.na
				highestva = nava.va
			}

		}

		// Determine if a higher proposal number was seen and adjust value
		ourValue := args.V

		if highestna > 0 {
			ourValue = highestva
		}

		// Setup accept phase with possibly new value
		aArgs := &paxosrpc.AcceptArgs{Key: args.Key, N: args.N, V: ourValue}
		replies := make(chan int, pn.numNodes)

		// Send out accept messages
		for i := 0; i < pn.numNodes; i++ {
			client = pn.allNodes[i]
			go sendAccept(pn, client, replies, aArgs)
		}

		// Receive accept messages
		for k := 0; k <= pn.numNodes/2; k++ {
			_ = <-replies
		}

		// Send commit messages
		cArgs := &paxosrpc.CommitArgs{Key: args.Key, V: ourValue}
		for i := 0; i < pn.numNodes; i++ {
			sendCommit(pn, i, cArgs)
		}
		// Notify main routine that we finished and commited ourValue
		doneChan <- ourValue
	}()

	// Run the main function and wait for up to 15 seconds
	select {
	case r := <-doneChan:
		// Propse completed successfully
		reply.V = r
		return nil
	case <-time.After(15 * time.Second):
		return nil
	}
}

func sendProposal(pn *paxosNode, client *rpc.Client, pArgs *paxosrpc.PrepareArgs, acceptChan chan Na_va) {
	var reply paxosrpc.PrepareReply
	client.Call("PaxosNode.RecvPrepare", pArgs, &reply)
	va := reply.V_a
	na := reply.N_a
	acceptChan <- Na_va{na: na, va: va}
	return
}

func sendAccept(pn *paxosNode, client *rpc.Client, replies chan int, aArgs *paxosrpc.AcceptArgs) {
	var reply paxosrpc.AcceptReply
	client.Call("PaxosNode.RecvAccept", aArgs, &reply)
	if reply.Status == paxosrpc.OK {
		replies <- 1
	}
	return
}

func sendCommit(pn *paxosNode, i int, cArgs *paxosrpc.CommitArgs) {
	client := pn.allNodes[i]
	var reply paxosrpc.CommitReply
	client.Call("PaxosNode.RecvCommit", cArgs, &reply)
	return
}

func (pn *paxosNode) GetValue(args *paxosrpc.GetValueArgs, reply *paxosrpc.GetValueReply) error {
	key := args.Key
	if value, found := pn.store[key]; found {
		reply.Status = paxosrpc.KeyFound
		reply.V = value
	} else {
		reply.Status = paxosrpc.KeyNotFound
	}
	return nil
}

func (pn *paxosNode) RecvPrepare(args *paxosrpc.PrepareArgs, reply *paxosrpc.PrepareReply) error {
	if _, ok := pn.highestSeen[args.Key]; !ok {
		pn.highestSeen[args.Key] = -1
	}
	if pn.highestSeen[args.Key] > args.N {
		reply.Status = paxosrpc.Reject
		reply.N_a = -1
		reply.V_a = nil

		return nil
	}
	if _, ok := pn.na[args.Key]; !ok {
		pn.na[args.Key] = -1
		pn.va[args.Key] = nil
	}
	pn.highestSeen[args.Key] = args.N
	reply.Status = paxosrpc.OK
	reply.N_a = pn.na[args.Key]
	reply.V_a = pn.va[args.Key]

	return nil
}

func (pn *paxosNode) RecvAccept(args *paxosrpc.AcceptArgs, reply *paxosrpc.AcceptReply) error {
	if _, ok := pn.highestSeen[args.Key]; !ok {
		pn.highestSeen[args.Key] = -1
	}

	if pn.highestSeen[args.Key] > args.N {
		reply.Status = paxosrpc.Reject
		return nil
	}
	pn.highestSeen[args.Key] = args.N
	pn.na[args.Key] = args.N
	pn.va[args.Key] = args.V
	reply.Status = paxosrpc.OK
	return nil
}

func (pn *paxosNode) RecvCommit(args *paxosrpc.CommitArgs, reply *paxosrpc.CommitReply) error {
	key := args.Key
	value := args.V
	pn.store[key] = value
	pn.na[key] = -1
	pn.va[key] = nil
	return nil
}

func (pn *paxosNode) RecvReplaceServer(args *paxosrpc.ReplaceServerArgs, reply *paxosrpc.ReplaceServerReply) error {
	client, _ := rpc.DialHTTP("tcp", args.Hostport)
	pn.allNodes[args.SrvID] = client
	pn.hostMap[args.SrvID] = args.Hostport
	return nil
}

func (pn *paxosNode) RecvReplaceCatchup(args *paxosrpc.ReplaceCatchupArgs, reply *paxosrpc.ReplaceCatchupReply) error {
	reply.Data, _ = json.Marshal(pn.store)
	return nil
}
