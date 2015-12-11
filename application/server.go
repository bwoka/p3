package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TodoServer struct {
	port string
}

type todoServer interface {
	addHandler(w http.ResponseWriter, r *http.Request)
	deleteHandler(w http.ResponseWriter, r *http.Request)
	renderPage(w http.ResponseWriter, r *http.Request)
	start(port string)
}

func getPaxos(key string) string {
	return "helloTest"
}

func setPaxos(key string, value string) {
}

func (ts *TodoServer) start(port string, done chan bool) {
	http.HandleFunc("/add", ts.addHandler)
	http.HandleFunc("/delete", ts.deleteHandler)
	http.HandleFunc("/", ts.renderPage)
	done <- true
	http.ListenAndServe(port, nil)

}

func (ts *TodoServer) renderPage(w http.ResponseWriter, r *http.Request) {
	header := "<!DOCTYPE html>\n<html>\n<body>\n<h1>Shared Todo List</h1>"
	//todo: probably have to adjust this one
	content := getPaxosContent(r.URL.Path)
	footer := "<form action=\"/add\" method=\"GET\">New Item:\n<input type=\"text\" name=\"newItem\"></input>\n<input type=\"submit\"></input></form>\n</body>\n</html>"
	fmt.Fprintf(w, strings.Join([]string{header, content, footer}, "\n"))
}

func addtoPaxos(s string) {
	last, _ := strconv.Atoi(getPaxos("topitem"))
	setPaxos(strconv.Itoa(last+1), s)
}

func deleteFromPaxos(s string) {
	setPaxos(s, "")
}

func getPaxosContent(url string) string {
	last, _ := strconv.Atoi(getPaxos("topitem"))
	content := make([]string, last+1)
	for x := 0; x <= last; x++ {
		xString := strconv.Itoa(x)
		item := getPaxos(xString)
		content[x] = strings.Join([]string{"<div><a href=\"", strings.Split(url, "/")[0], "delete?toDelete=", xString, "\">X  </a> ", item, "</div>\n"}, "")
	}
	return strings.Join(content, "")
}

func (ts *TodoServer) deleteHandler(w http.ResponseWriter, r *http.Request) {
	toDelete := r.Header.Get("toDelete")
	deleteFromPaxos(toDelete)
	ts.renderPage(w, r)
}

func (ts *TodoServer) addHandler(w http.ResponseWriter, r *http.Request) {
	newItem := r.Header.Get("newItem")
	addtoPaxos(newItem)
	ts.renderPage(w, r)
}

func main() {
	fmt.Println("hello")
	done := make(chan bool)
	setPaxos("topitem", "0")

	server1 := TodoServer{port: "8080"}
	go server1.start(":8080", done)
	_ = <-done
	server2 := TodoServer{port: "8081"}
	go server2.start(":8081", done)
	_ = <-done
	//server3:=TodoServer{port: "8082"}
	//go server3.start(":8082")
	//_=<-done
	//server4:=TodoServer{port: "8083"}
	//go server4.start(":8083")
	//_=<-done
	time.Sleep(1)

}
