package main

import (
	"flag"
	"fmt"
	"github.com/cmu440-F15/paxosapp/paxos"
	"github.com/cmu440-F15/paxosapp/rpc/paxosrpc"
	"net/http"
	"strconv"
	"strings"
)

type TodoServer struct {
	port string
	node paxos.PaxosNode
}

type todoServer interface {
	addHandler(w http.ResponseWriter, r *http.Request)
	deleteHandler(w http.ResponseWriter, r *http.Request)
	renderPage(w http.ResponseWriter, r *http.Request)
	start(port string)
	getPaxos(key string) string
	setPaxos(key string, value string)
	deleteFromPaxos(s string)
	addToPaxos(s string)
}

func (ts *TodoServer) getPaxos(key string) string {

	gArgs := &paxosrpc.GetValueArgs{Key: key}
	var reply paxosrpc.GetValueReply
	ts.node.GetValue(gArgs, &reply)

	res := reply.V
	if str, ok := res.(string); ok {
		return str
	}

	return ""

}

func (ts *TodoServer) setPaxos(key string, value string) {
	set := false
	for set == false {
		nArgs := &paxosrpc.ProposalNumberArgs{Key: key}
		var nreply paxosrpc.ProposalNumberReply
		ts.node.GetNextProposalNumber(nArgs, &nreply)
		pArgs := &paxosrpc.ProposeArgs{Key: key, V: value, N: nreply.N}
		var preply paxosrpc.ProposeReply
		ts.node.Propose(pArgs, &preply)
		if _, ok := preply.V.(string); ok {
			set = true
		}
	}
	return
}

func (ts *TodoServer) start(port string) {

	http.HandleFunc("/add", ts.addHandler)
	http.HandleFunc("/delete", ts.deleteHandler)
	http.HandleFunc("/", ts.renderPage)
	http.ListenAndServe(port, nil)

}

func (ts *TodoServer) renderPage(w http.ResponseWriter, r *http.Request) {
	header := "<!DOCTYPE html>\n<html>\n<body>\n<h1>Shared Todo List</h1>"
	//todo: probably have to adjust this one
	content := getPaxosContent(r.URL.Path, ts)
	footer := "<form action=\"/add\" method=\"GET\">New Item:\n<input type=\"text\" name=\"newItem\"></input>\n<input type=\"submit\"></input></form>\n</body>\n</html>"
	fmt.Fprintf(w, strings.Join([]string{header, content, footer}, "\n"))
}

func (ts *TodoServer) addtoPaxos(s string) {
	last, _ := strconv.Atoi(ts.getPaxos("topitem"))
	ts.setPaxos("topitem", strconv.Itoa(last+1))
	ts.setPaxos(strconv.Itoa(last), s)
}

func (ts *TodoServer) deleteFromPaxos(s string) {
	ts.setPaxos(s, "")
}

func getPaxosContent(url string, ts *TodoServer) string {
	last, _ := strconv.Atoi(ts.getPaxos("topitem"))
	content := make([]string, last+1)
	for x := 0; x < last; x++ {
		xString := strconv.Itoa(x)
		item := ts.getPaxos(xString)
		if item != "" {
			content[x] = strings.Join([]string{"<div><a href=\"", strings.Split(url, "/")[0], "delete?toDelete=", xString, "\">X  </a> ", item, "</div>\n"}, "")
		}
	}
	return strings.Join(content, "")
}

func (ts *TodoServer) deleteHandler(w http.ResponseWriter, r *http.Request) {
	lastArg := strings.Split(r.URL.String(), "=")
	toDelete := lastArg[len(lastArg)-1]
	ts.deleteFromPaxos(toDelete)
	ts.renderPage(w, r)
}

func (ts *TodoServer) addHandler(w http.ResponseWriter, r *http.Request) {
	lastArg := strings.Split(r.URL.String(), "=")
	newItem := lastArg[len(lastArg)-1]
	ts.addtoPaxos(newItem)
	ts.renderPage(w, r)
}

func main() {
	hmap := map[int]string{
		0: ":8010",
		1: ":8012",
		2: ":8013",
		3: ":8015"}

	num := flag.Int("num", 0, "the number of the paxos node for this server")
	port := flag.String("port", "foo", "a string")

	flag.Parse()

	node, _ := paxos.NewPaxosNode(hmap[*num], hmap, 4, *num, 4, false)
	server := &TodoServer{port: *port, node: node}

	server.setPaxos("topitem", "0")
	server.start(strings.Join([]string{":", *port}, ""))

}
