package drpc

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Server struct {
	stopped   bool
	OnConnect func(*Connection)
	OnClose   func(string)

	apiHandlers map[string]ApiHandler
	connections map[string]*Connection
}

func NewServer() *Server {
	s := &Server{
		apiHandlers: make(map[string]ApiHandler),
		connections: make(map[string]*Connection),
	}
	return s
}

func (me *Server) client(id string) *Connection {
	return me.connections[id]
}

func (me *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if me.stopped {
		return
	}

	ws, err := websocket.Upgrade(w, req, nil, 1024, 1024)
	if err != nil {
		log.Println("drpc-error: websocket-upgrade", err)
		return
	}

	me.wshandler(req, ws)
}

func (me *Server) wshandler(req *http.Request, ws *websocket.Conn) {
	ep := newConnection()
	ep.apiHandlers = me.apiHandlers
	ep.conn = ws

	id := ep.Id()
	me.connections[id] = ep
	defer me.onClose(id)

	if me.OnConnect != nil {
		go me.OnConnect(ep)
	}

	ep.workloop()
}

func (me *Server) onClose(id string) {
	delete(me.connections, id)
	if me.OnClose != nil {
		me.OnClose(id)
	}
}

func (me *Server) Handle(cmd string, f ApiHandler) {
	me.apiHandlers[cmd] = f
}
