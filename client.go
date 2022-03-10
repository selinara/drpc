package drpc

import (
	"crypto/tls"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	ep        *Connection
	Id        string
	OnConnect func()
	OnClose   func()
	Online    bool
	last_err  string
}

func NewClient(id string) (c *Client) {
	c = &Client{Id: id}
	c.ep = newConnection()
	return c
}

func (me *Client) Channel() *Channel {
	if me.ep == nil {
		return nil
	}
	return me.ep.Channel()
}

func (me *Client) Connect(addr string) (err error) {
	ch := make(chan error)
	go me.loop(addr, ch)
	err = <-ch
	return err
}

func (me *Client) loop(addr string, ch chan error) {
	for {
		me.run(addr, ch)
		time.Sleep(time.Second)
	}
	return
}

func (me *Client) run(addr string, ch chan error) {
	var (
		use_ssl bool
		u       *url.URL
		c       net.Conn
		rsp     *http.Response
		conn    *websocket.Conn
	)

	u, err := url.Parse(addr)
	if err != nil {
		go func() {
			ch <- err
		}()
		me.log(err)
		return
	}

	switch u.Scheme {
	case "http":
	case "ws":
		use_ssl = false
	case "https":
	case "wss":
		use_ssl = true
	}
	c, err = net.Dial("tcp", u.Host)
	if err != nil {
		go func() {
			ch <- err
		}()
		me.log(err)
		return
	}

	if use_ssl {
		c = tls.Client(c, &tls.Config{InsecureSkipVerify: true})
	}

	headers := http.Header{}
	conn, rsp, err = websocket.NewClient(c, u, headers, 1024, 1024)
	if err != nil {
		go func() {
			ch <- err
		}()
		me.log(err, rsp)
		return
	}

	me.ep.conn = conn

	defer func() {
		me.Online = false
		if me.OnClose != nil {
			me.OnClose()
		}
	}()

	me.Online = true
	go func() {
		ch <- nil
	}()
	if me.OnConnect != nil {
		go me.OnConnect()
	}

	me.ep.workloop()
	return
}

func (me *Client) log(err error, v ...interface{}) {
	if me.last_err != err.Error() {
		log.Println(err)
	}
	me.last_err = err.Error()
}

func (me *Client) Handle(cmd string, f ApiHandler) {
	me.ep.handle(cmd, f)
}
