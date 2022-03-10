package drpc

import (
	"encoding/json"
	"fmt"
	"log"
)

const (
	actRequest  = 1
	actResponse = 2
	actPing     = 8
	actClose    = 9
)

type bag struct {
	Cid  string
	Act  int8
	Cmd  string
	Data []byte
	Args [][]byte
	Err  string
}

type Request struct {
	bag        *bag
	Connection *Connection
}

func (me *Request) UnmarshalArgs(args ...interface{}) (err error) {
	ra_len := len(me.bag.Args)
	for i, _ := range args {
		if ra_len > i {
			err = json.Unmarshal(me.bag.Args[i], args[i])
			if err != nil {
				return
			}
		}
	}
	return
}

type Response struct {
	Connection *Connection
	Data       interface{}
	Err        error
	bag        *bag
}

func (me *Response) tobag(req *bag) bag {
	var err error
	b := bag{
		Cid: req.Cid,
		Act: actResponse,
		Cmd: req.Cmd,
	}
	if me.Err != nil {
		b.Err = me.Err.Error()
	}
	b.Data, err = json.Marshal(me.Data)
	if err != nil {
		log.Println("drpc:", err)
		b.Err = err.Error()
	}
	return b
}

func (me *Response) Unmarshal(v interface{}) (err error) {
	err = json.Unmarshal(me.bag.Data, v)
	if me.Err != nil {
		err = me.Err
	}
	return
}

func (me *Response) Print() {
	fmt.Println(me.JsonString())
}

func (me *Response) JsonString() string {
	var v interface{}
	json.Unmarshal(me.bag.Data, &v)
	buf, _ := json.MarshalIndent(v, "", "    ")
	return string(buf)
}
