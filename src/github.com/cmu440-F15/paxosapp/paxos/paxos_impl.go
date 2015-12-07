package paxos

import (
	"errors"
	//	"fmt"
	"github.com/cmu440-F15/paxosapp/rpc/paxosrpc"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type paxosNode struct {
	// TODO: implement this!
	store       map[string]interface{}
	highestSeen map[string]int
	highestSeen map[string]int
	va map[string]interface{}
	na map[string]int

	nodeID      int
	hostMap     map[int]string
	propAcks    map[string]int
	acceptAcks  map[string]int
	myHostPort  string
	allNodes    []*rpc.Client
	numNodes    int

	stage map[string]int
}

type Na_va struct{
	na int
	va string
}

// NewPaxosNode creates a new PaxosNode. This function should return only when
// all nodes have joined the ring, and should return a non-nil error if the node
// could not be started in spite of dialing the other nodes numRetries times.
//
// hostMap is a map from node IDs to their hostports, numNodes is the number
// of nodes in the ring, replace is a flag which indicates whether this node
// is a replacement for a node which failed.
func NewPaxosNode(myHostPort string, hostMap map[int]string, numNodes, srvId, numRetries int, replace bool) (PaxosNode, error) {
	newNode := paxosNode{
		store:       make(map[string]interface{}),
		highestSeen: make(map[string]int),
		nodeID:      srvId,
		hostMap:     hostMap,
		propAcks:    make(map[string]int),
		acceptAcks:  make(map[string]int),
		myHostPort:  myHostPort,
		allNodes:    make([]*rpc.Client, numNodes),
		numNodes:    numNodes,
		stage:       make(map[string]int)}

	rpc.RegisterName("PaxosNode", &newNode)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", myHostPort)
	if e != nil {
		return nil, errors.New("PaxosNode couldn't start listening")
	}
	go http.Serve(l, nil)

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
				newNode.allNodes[i] = client
			}

		}

	}
	return &newNode, nil
	//	fmt.Println("woo")
	//	return nil, nil
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
	pArgs := &paxosrpc.PrepareArgs{Key: args.Key, N: args.N}
	var client *rpc.Client
	replies := make(chan int, pn.numNodes)
	acceptChan:=make(chan Na_va, pn.numNodes)
	for i := 0; i < pn.numNodes; i++ {
		client = pn.allNodes[i]
		go sendProposal(pn, client, replies, pArgs)
	}
	ackd := false
	var j int
	for j = 0; j < 1500; j++ {
		if len(replies) > pn.numNodes/2 {
			ackd = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	highestna:=0
	highestva:=nil
	close(acceptChan)
	for nava:= range acceptChan{
		if nava.na>=highestna{
			highestna=nava.na
			highestva=nava.va
		}
	}
	ourValue:= args.V

	if highestna!=nil{
		ourValue=highestva
	}

	if ackd == false {
		return nil
	}

	aArgs := &paxosrpc.AcceptArgs{Key: args.Key, N: args.N, V: ourValue}
	replies2 := make(chan int, pn.numNodes)
	for i := 0; i < pn.numNodes; i++ {
		client = pn.allNodes[i]
		go sendAccept(pn, client, replies2, aArgs)
	}
	ackd = false
	for k := j; k < 1500; k++ {
		if len(replies2) > pn.numNodes/2 {
			ackd = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if ackd == false {
		return nil
	}

	cArgs := &paxosrpc.CommitArgs{Key: args.Key, V: outValue}
	for i := 0; i < pn.numNodes; i++ {
		sendCommit(pn, i, cArgs)
	}
	
	reply.V = args.V
	return nil
}

func sendProposal(pn *paxosNode, client *rpc.Client, replies chan int, pArgs *paxosrpc.PrepareArgs, acceptChan chan Na_va) {
	var reply paxosrpc.PrepareReply
	client.Call("PaxosNode.RecvPrepare", pArgs, &reply)
	acceptChan <- Na_va{na=reply.n_a, va=reply.va}
	replies <- 1
	return 
}

func sendAccept(pn *paxosNode, client *rpc.Client, replies chan int, aArgs *paxosrpc.AcceptArgs) {
	var reply paxosrpc.AcceptReply
	client.Call("PaxosNode.RecvAccept", aArgs, &reply)
	replies <- 1
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
	
	if highestSeen[args.Key]>args.N{
		reply.Status=paxosrpc.Reject
		reply.N_a=-1
		reply.V_a=nil

		return nil
	}
	highestSeen[args.Key]=args.N
	reply.Status = paxosrpc.OK
	reply.N_a=na[args.Key]
	reply.V_a=va[args.Key]

	return nil
}

func (pn *paxosNode) RecvAccept(args *paxosrpc.AcceptArgs, reply *paxosrpc.AcceptReply) error {
	
	if highestSeen[args.Key]>args.N{
		reply.Status=paxosrpc.Reject
		return nil
	}
	highestSeen[args.Key]=args.N
	na[args.Key]=args.N
	va[args.Key]=args.N
	reply.Status = paxosrpc.OK
	return nil
}

func (pn *paxosNode) RecvCommit(args *paxosrpc.CommitArgs, reply *paxosrpc.CommitReply) error {
	key := args.Key
	value := args.V
	pn.store[key] = value
	return nil
}

func (pn *paxosNode) RecvReplaceServer(args *paxosrpc.ReplaceServerArgs, reply *paxosrpc.ReplaceServerReply) error {
	return errors.New("not implemented")
}

func (pn *paxosNode) RecvReplaceCatchup(args *paxosrpc.ReplaceCatchupArgs, reply *paxosrpc.ReplaceCatchupReply) error {
	return errors.New("not implemented")
}
