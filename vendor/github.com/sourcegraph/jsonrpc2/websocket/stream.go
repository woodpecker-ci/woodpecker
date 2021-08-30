// Package websocket provides WebSocket transport support for JSON-RPC
// 2.0.
package websocket

import (
	"io"

	ws "github.com/gorilla/websocket"
)

// A ObjectStream is a jsonrpc2.ObjectStream that uses a WebSocket to
// send and receive JSON-RPC 2.0 objects.
type ObjectStream struct {
	conn *ws.Conn
}

// NewObjectStream creates a new jsonrpc2.ObjectStream for sending and
// receiving JSON-RPC 2.0 objects over a WebSocket.
func NewObjectStream(conn *ws.Conn) ObjectStream {
	return ObjectStream{conn: conn}
}

// WriteObject implements jsonrpc2.ObjectStream.
func (t ObjectStream) WriteObject(obj interface{}) error {
	return t.conn.WriteJSON(obj)
}

// ReadObject implements jsonrpc2.ObjectStream.
func (t ObjectStream) ReadObject(v interface{}) error {
	err := t.conn.ReadJSON(v)
	if e, ok := err.(*ws.CloseError); ok {
		if e.Code == ws.CloseAbnormalClosure && e.Text == io.ErrUnexpectedEOF.Error() {
			// Suppress a noisy (but harmless) log message by
			// unwrapping this error.
			err = io.ErrUnexpectedEOF
		}
	}
	return err
}

// Close implements jsonrpc2.ObjectStream.
func (t ObjectStream) Close() error {
	return t.conn.Close()
}
