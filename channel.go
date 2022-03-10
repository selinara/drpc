package drpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

type Channel struct {
	id     string
	leader *Connection
	lk     sync.Mutex
	ch     chan *bag
}

func (me *Channel) Call(cmd string, args ...interface{}) (rsp *Response) {
	var err error
	me.lk.Lock()
	me.lk.Unlock()

	b := bag{
		Act:  actRequest,
		Cid:  me.Id(),
		Cmd:  cmd,
		Args: make([][]byte, len(args)),
	}

	for i, arg := range args {
		b.Args[i], _ = json.Marshal(arg)
	}

	rsp = &Response{Connection: me.leader}

	if err != nil {
		rsp.Err = err
		return
	}

	err = me.leader.send(b)
	if err != nil {
		rsp.Err = err
		return
	}

	select {
	case result := <-me.ch:
		rsp.bag = result
		if result.Err != "" {
			rsp.Err = errors.New(result.Err)
		}
		return
	}
	return
}

func (me *Channel) Id() string {
	if me.id == "" {
		me.id = fmt.Sprintf("%p", me)
	}
	return me.id
}

func (me *Channel) Close() {
	me.leader.closeChannel(me.id)
}
