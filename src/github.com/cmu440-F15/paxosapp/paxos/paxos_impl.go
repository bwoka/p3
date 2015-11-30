package paxos

import (
	"errors"
	"github.com/cmu440-F15/paxosapp/rpc/paxosrpc"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type paxosNode struct {
	// TODO: implement this!
	store       map[string]string
	highestSeen map[string]int
	nodeID      int
	hostMap     map[int]string
	propAcks    map[string]int
	acceptAcks  map[string]int
	myHostPort  string

	allNodes []*rpc.Client

	stage map[string]int
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
		store:       make(map[string]string),
		highestSeen: make(map[string]int),
		nodeID:      srvId,
		hostMap:     hostMap,
		propAcks:    make(map[string]int),
		acceptAcks:  make(map[string]int),
		myHostPort:  myHostPort,
		allNodes:    make([]*rpc.Client, numNodes),
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

}

func (pn *paxosNode) GetNextProposalNumber(args *paxosrpc.ProposalNumberArgs, reply *paxosrpc.ProposalNumberReply) error {
	return errors.New("not implemented")
}

func (pn *paxosNode) Propose(args *paxosrpc.ProposeArgs, reply *paxosrpc.ProposeReply) error {
	return errors.New("not implemented")
}

func (pn *paxosNode) GetValue(args *paxosrpc.GetValueArgs, reply *paxosrpc.GetValueReply) error {
	return errors.New("not implemented")
}

func (pn *paxosNode) RecvPrepare(args *paxosrpc.PrepareArgs, reply *paxosrpc.PrepareReply) error {
	return errors.New("not implemented")
}

func (pn *paxosNode) RecvAccept(args *paxosrpc.AcceptArgs, reply *paxosrpc.AcceptReply) error {
	return errors.New("not implemented")
}

func (pn *paxosNode) RecvCommit(args *paxosrpc.CommitArgs, reply *paxosrpc.CommitReply) error {
	return errors.New("not implemented")
}

func (pn *paxosNode) RecvReplaceServer(args *paxosrpc.ReplaceServerArgs, reply *paxosrpc.ReplaceServerReply) error {
	return errors.New("not implemented")
}

func (pn *paxosNode) RecvReplaceCatchup(args *paxosrpc.ReplaceCatchupArgs, reply *paxosrpc.ReplaceCatchupReply) error {
	return errors.New("not implemented")
}
