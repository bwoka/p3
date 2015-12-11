package main

import (
    "fmt"
    "net/http"
    "strconv"
    "strings"
)

type TodoServer struct{
    port string}

type todoServer interface {
    addHandler(w http.ResponseWriter, r *http.Request)
    deleteHandler(w http.ResponseWriter, r *http.Request)
    renderPage(w http.ResponseWriter, r *http.Request)
    start(port string)
}


func getPaxos(key string){

}

func setPaxos(key string,value string){

}

func (ts *TodoServer ) start(port string) {
    http.HandleFunc("/add", ts.addHandler)
    http.HandleFunc("/delete", ts.deleteHandler)
    http.HandleFunc("/", ts.renderPage)
    http.ListenAndServe(port, nil)
}


func (ts *TodoServer ) renderPage(w http.ResponseWriter, r *http.Request) {
     header:="<!DOCTYPE html>\n<html>\n<body>\n<h1>Shared Todo List</h1>"
     //todo: probably have to adjust this one
     content:=getPaxosContent(r.URL.Path)
     footer:="<form action=\"/delete\" method=\"GET\">New Item:<input type=\"text\" name=\"newItem\"></form>\n</body>\n</html>"
     fmt.Fprintf(w, strings.Join([]string{header, content,footer},"\n"), r.URL.Path[1:])
}


func addtoPaxos(s string){
    last,_:= strconv.Atoi(getPaxos("topitem"))
    setPaxos(strconv.Itoa(last+1),s)
}

func deleteFromPaxos(s string){
    setPaxos(s,"")
}
    



func getPaxosContent(url string) string{
    last,_:= strconv.Atoi(getPaxos("topitem"))
    content:=make([]string, last+1)
    for x:=0; x<=last; x++{
        xString:=strconv.Itoa(x)
        item:=getPaxos(xString)
        content[x]=strings.Join([]string{"<div><a href=\"",url,"delete?name1=value1&name2=value2\">",item,"</div>\n"},"")
    }
    return strings.Join(content,"")
}

func (ts *TodoServer )  deleteHandler(w http.ResponseWriter, r *http.Request) {
    toDelete := r.Header.Get("toDelete")
    deleteFromPaxos(toDelete)
    ts.renderPage(w, r)
}

func (ts *TodoServer ) addHandler(w http.ResponseWriter, r *http.Request) {
    newItem := r.Header.Get("newItem")
    addtoPaxos(newItem)
    ts.renderPage(w, r)
}


func main(){
    server1:=TodoServer{port: "8080"}
    go server1.start(":8080")
    server2:=TodoServer{port: "8081"}
    go server2.start(":8081")
    server3:=TodoServer{port: "8082"}
    go server3.start(":8082")
    server4:=TodoServer{port: "8083"}
    go server4.start(":8083")
    setPaxos("topitem","0")
}
