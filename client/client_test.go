package client_test

import (
	//	"log"
	proto "github.com/percona/cloud-protocol"
	"github.com/percona/cloud-tools/client"
	"github.com/percona/cloud-tools/test"
	"github.com/percona/cloud-tools/test/mock"
	. "launchpad.net/gocheck"
	"testing"
	"time"
)

// Hook gocheck into the "go test" runner.
// http://labix.org/gocheck
func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
	server *mock.WebsocketServer
}

var _ = Suite(&TestSuite{})

const (
	ADDR     = "127.0.0.1:8000" // make sure this port is free
	URL      = "ws://" + ADDR
	ENDPOINT = "/"
)

func (s *TestSuite) SetUpSuite(t *C) {
	mock.SendChan = make(chan interface{}, 5)
	mock.RecvChan = make(chan interface{}, 5)
	s.server = new(mock.WebsocketServer)
	go s.server.Run(ADDR, ENDPOINT)
	time.Sleep(100 * time.Millisecond)
}

func (s *TestSuite) TearDownTest(t *C) {
	for _, c := range mock.Clients {
		mock.ClientDisconnectChan <- c
	}
}

/////////////////////////////////////////////////////////////////////////////
// Test cases
// //////////////////////////////////////////////////////////////////////////

func (s *TestSuite) TestSend(t *C) {
	origin := "http://localhost:1"
	ws, err := client.NewWebsocketClient(URL+ENDPOINT, origin)
	t.Assert(err, IsNil)

	err = ws.Connect()
	t.Assert(err, IsNil)

	logEntry := &proto.LogEntry{
		Level:   2,
		Service: "qan",
		Msg:     "Hello",
	}
	err = ws.Send(logEntry)
	t.Assert(err, IsNil)

	// todo: this is probably prone to deadlocks, not thread-safe
	c, ok := mock.Clients[origin]
	if !t.Check(ok, Equals, true) {
		return
	}

	got := test.WaitData(c.RecvChan)
	if !t.Check(len(got), Equals, 1) {
		return
	}
	// We're dealing with generic data; see
	// http://blog.golang.org/json-and-go
	m := got[0].(map[string]interface{})
	t.Assert(m["Level"], Equals, float64(2))
	t.Assert(m["Service"], Equals, "qan")
	t.Assert(m["Msg"], Equals, "Hello")

	ws.Disconnect()
	t.Assert(err, IsNil)

	// todo: handle this better
	time.Sleep(100 * time.Millisecond) // yield thread
	_, ok = mock.Clients[origin]
	t.Assert(ok, Equals, false)
}

// Test channel-based interface.
func (s *TestSuite) TestChannels(t *C) {
	origin := "http://localhost:2"
	ws, err := client.NewWebsocketClient(URL+ENDPOINT, origin)
	t.Assert(err, IsNil)

	err = ws.Connect()
	t.Assert(err, IsNil)

	// todo: stop the threads
	go ws.Run()

	// todo: this is probably prone to deadlocks, not thread-safe
	c, ok := mock.Clients[origin]
	if !t.Check(ok, Equals, true) {
		return
	}

	// API sends Cmd to client.
	cmd := &proto.Cmd{
		User: "daniel",
		Ts:   time.Now(),
		Cmd:  "Status",
	}
	c.SendChan <- cmd

	// If client's recvChan is working, it will receive the Cmd.
	got := test.WaitCmd(ws.RecvChan())
	if !t.Check(len(got), Equals, 1) {
		return
	}
	t.Assert(got[0], DeepEquals, *cmd)

	// Client sends Reply in response to Cmd.
	reply := cmd.Reply("", nil)
	ws.SendChan() <- reply

	// If client's sendChan is working, we/API will receive the Reply.
	data := test.WaitData(c.RecvChan)
	if !t.Check(len(data), Equals, 1) {
		return
	}
	// We're dealing with generic data again.
	m := data[0].(map[string]interface{})
	t.Assert(m["Cmd"], Equals, "Status")
	t.Assert(m["Error"], Equals, "")

	ws.Disconnect()
	t.Assert(err, IsNil)
}

func (s *TestSuite) TestApiDisconnect(t *C) {
	origin := "http://localhost:3"
	ws, err := client.NewWebsocketClient(URL+ENDPOINT, origin)
	t.Assert(err, IsNil)

	err = ws.Connect()
	t.Assert(err, IsNil)

	// todo: this is probably prone to deadlocks, not thread-safe
	c, ok := mock.Clients[origin]
	if !t.Check(ok, Equals, true) {
		return
	}

	// No error yet.
	got := test.WaitErr(ws.ErrorChan())
	t.Assert(len(got), Equals, 0)

	mock.ClientDisconnectChan <- c

	/**
	 * I cannot provoke an error on websocket.Send(), only Receive().
	 * Perhaps errors (e.g. ws closed) are only reported on recv?
	 * This only affect the logger since it's ws send-only: it will
	 * need a goroutine blocking on Recieve() that, upon error, notifies
	 * the sending goroutine to reconnect.
	 */
	var data interface{}
	err = ws.Recv(data)
	t.Assert(err, NotNil) // EOF due to disconnect.
}

func (s *TestSuite) TestApiDisconnectChan(t *C) {
	origin := "http://localhost:4"
	ws, err := client.NewWebsocketClient(URL+ENDPOINT, origin)
	t.Assert(err, IsNil)

	err = ws.Connect()
	t.Assert(err, IsNil)

	go ws.Run()

	// todo: this is probably prone to deadlocks, not thread-safe
	c, ok := mock.Clients[origin]
	if !t.Check(ok, Equals, true) {
		return
	}

	// No error yet.
	got := test.WaitErr(ws.ErrorChan())
	t.Assert(len(got), Equals, 0)

	// API sends Cmd to client.
	cmd := &proto.Cmd{
		User: "daniel",
		Ts:   time.Now(),
		Cmd:  "Status",
	}
	c.SendChan <- cmd

	// No error yet.
	got = test.WaitErr(ws.ErrorChan())
	t.Assert(len(got), Equals, 0)

	mock.ClientDisconnectChan <- c

	got = test.WaitErr(ws.ErrorChan())
	if !t.Check(len(got), Equals, 1) {
		return
	}
	t.Assert(got[0], NotNil)
}