README.md

Server

```
func main(){
    rpc_server := drpc.NewServer()

    rpc_server.Handle("hello", func_hello)

    httpd := http.NewServeMux()
    httpd.Handle("/api/rpc", rpc_server)

    rpc_server.OnConnect = on_connect
    rpc_server.OnClose = on_close

    addr := ":8010"
    err := http.ListenAndServe(addr, httpd)
    if err != nil {
        panic(err)
    } 
}

func func_hello(r *drpc.Request) drpc.Response {

    var v string
    err := r.UnmarshalArgs(&v)
    if err != nil {
        log.Println("error:", err)
    }

    return drpc.Response{
        Data: "world",
    }
}

func on_connect(conn *drpc.Connection) {
    log.Println("connect", conn.Id())

    c := conn.Channel()
    defer c.Close()

    rsp := c.Call("remote")
    if rsp.Err != nil {
        log.Println(111, rsp.Err)
    }
    rsp.Print()
}

func on_close(conn_id string) {
    log.Println("closed ", conn_id)
}
```

--------------------------

Client

```
func main(){
    c := drpc.NewClient()
    c.Handle("remote", func_remote)
        
    err := c.Connect("http://127.0.0.1:8010/api/rpc")
    if err != nil {
        panic(err)
    }
    ch := c.Channel()

    var s string
    err = ch.Call("hello").Unmarshal(&s)
    fmt.Println("hello", s, "errors=", err)   
}

func func_remote(r *drpc.Request) drpc.Response {
    fmt.Println("on remote call")
    return drpc.Response{Data: "hello"}
}    
```