
import (
    "fmt"
    "net/http"
    "strconv"
)

struct TodoServer{
    port string
}

func main(){
    server1:=ts{port: "8080"}
    go server1.start(":8080")
    server2:=ts{port: "8081"}
    go server2.start(":8081")
    server3:=ts{port: "8082"}
    go server3.start(":8082")
    server4:=ts{port: "8083"}
    go server4.start(":8083")
    paxos.add("topitem","0")
}

func (*todoServer ts) start(port string) {
    http.HandleFunc("/add", ts.addHandler)
    http.HandleFunc("/delete", ts.deleteHandler)
    http.HandleFunc("/", ts.renderPage)
    http.ListenAndServe(port, nil)
}

func getPaxosContent(url string)[]string{
    last:= strconv.Atoi(getPaxos("topitem"))
    content:=make([]sting, last+1)
    for x:=0; x<=last; x++{
        xString=strconv.Itoa(x)
        item:=getPaxos(xString)
        content[x]=strings.Join([]string{'<div><a href="',url,'delete?name1=value1&name2=value2">',item,"</div>\n"},"")
    }
    return strings.Join(content,"")
}

func (*todoServer ts)  deleteHandler(w http.ResponseWriter, r *http.Request) {
    toDelete := r.Header.Get("toDelete")
    deleteFromPaxos(toDelete)
    renderPage(w http.ResponseWriter, r *http.Request)
}

func (*todoServer ts) addHandler(w http.ResponseWriter, r *http.Request) {
    toDelete := r.Header.Get("newItem")
    addtoPaxos(newItem)
    renderPage(w http.ResponseWriter, r *http.Request)
}

func (*todoServer ts) renderPage(w http.ResponseWriter, r *http.Request) {
     header:="<!DOCTYPE html>\n<html>\n<body>\n<h1>Shared Todo List</h1>"
     //todo: probably have to adjust this one
     content:=getPaxosContent(r.Url.Path[0])
     footer:='<form action="/delete" method="GET">New Item:<input type="text" name="newItem"></form>\n</body>\n</html>'
     fmt.Fprintf(w, strings.Join([]string{header, content,footer},"\n"), r.URL.Path[1:])
}
