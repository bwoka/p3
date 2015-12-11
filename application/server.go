
import (
    "fmt"
    "net/http"
)

struct TodoServer{
}

func main(){
    server1:=ts{}
    go server1.start(":8080")
    server2:=ts{}
    go server2.start(":8081")
    server3:=ts{}
    go server3.start(":8082")
    server4:=ts{}
    go server4.start(":8083")
    paxos.add("topitem","0")
}

func (*todoServer ts) start(port string) {
    http.HandleFunc("/add", ts.addHandler)
    http.HandleFunc("/delete", ts.deleteHandler)
    http.HandleFunc("/", ts.renderPage)
    http.ListenAndServe(port, nil)
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
     content:=getPaxosContent()
     footer:="</body>\n</html>"
     fmt.Fprintf(w, strings.Join([header, content,footer],"\n"), r.URL.Path[1:])
}


