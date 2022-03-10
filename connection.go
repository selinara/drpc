package drpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"sync"
	"time"
)

type Connection struct {
	id          string
	client_id   string
	connectAt   time.Time
	apiHandlers map[string]ApiHandler
	channels    map[string]*Channel
	lk          sync.Mutex
	conn        *websocket.Conn
}

type ApiHandler func(*Request) Response

func (me *Connection) Id() string {
	if me.id == "" {
		me.id = fmt.Sprintf("%p", me)
	}
	return me.id
}

func (me *Connection) LocalAddr() net.Addr {
	return me.conn.UnderlyingConn().LocalAddr()
}

func (me *Connection) RemoteAddr() net.Addr {
	return me.conn.UnderlyingConn().RemoteAddr()
}

func (me *Connection) Channel() *Channel {
	c := &Channel{leader: me, ch: make(chan *bag)}
	me.channels[c.Id()] = c
	return c
}

func (me *Connection) closeChannel(id string) {
	delete(me.channels, id)
}

func (me *Connection) send(v bag) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return me.conn.WriteMessage(1, buf)
}

func (me *Connection) workloop() {
	for {
		v := &bag{}
		_, p, err := me.conn.ReadMessage()

		if err != nil {
			log.Println(err)
			return
		}
		err = json.Unmarshal(p, v)
		if err != nil {
			log.Println(err)
			return
		}
		switch v.Act {
		case actRequest:
			me.handle_request(v)
		case actResponse:
			me.handle_result(v)
		case actClose:
			log.Println("closed")
			return
		}
	}
}

func (me *Connection) handle_request(v *bag) {
	f, ok := me.apiHandlers[v.Cmd]
	var rsp Response
	if ok {
		rsp = f(&Request{bag: v, Connection: me})
	} else {
		rsp = Response{Err: errors.New("api notfound")}
	}
	me.conn.WriteJSON(rsp.tobag(v))
}

func (me *Connection) handle_result(v *bag) {
	c, ok := me.channels[v.Cid]
	if ok {
		c.ch <- v
	}
}

func newConnection() *Connection {
	return &Connection{
		apiHandlers: make(map[string]ApiHandler),
		channels:    make(map[string]*Channel),
		connectAt:   time.Now(),
	}
}

func (me *Connection) handle(cmd string, f ApiHandler) {
	me.apiHandlers[cmd] = f
}
