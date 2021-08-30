package jsonrpc2

import (
	"encoding/json"
	"sync"
)

// Logger interface implements one method - Printf.
// You can use the stdlib logger *log.Logger
type Logger interface {
	Printf(format string, v ...interface{})
}

// ConnOpt is the type of function that can be passed to NewConn to
// customize the Conn before it is created.
type ConnOpt func(*Conn)

// OnRecv causes all requests received on conn to invoke f(req, nil)
// and all responses to invoke f(req, resp),
func OnRecv(f func(*Request, *Response)) ConnOpt {
	return func(c *Conn) { c.onRecv = append(c.onRecv, f) }
}

// OnSend causes all requests sent on conn to invoke f(req, nil) and
// all responses to invoke f(nil, resp),
func OnSend(f func(*Request, *Response)) ConnOpt {
	return func(c *Conn) { c.onSend = append(c.onSend, f) }
}

// LogMessages causes all messages sent and received on conn to be
// logged using the provided logger.
func LogMessages(logger Logger) ConnOpt {
	return func(c *Conn) {
		// Remember reqs we have received so we can helpfully show the
		// request method in OnSend for responses.
		var (
			mu         sync.Mutex
			reqMethods = map[ID]string{}
		)

		// Set custom logger from provided input
		c.logger = logger

		OnRecv(func(req *Request, resp *Response) {
			switch {
			case req != nil:
				mu.Lock()
				reqMethods[req.ID] = req.Method
				mu.Unlock()

				params, _ := json.Marshal(req.Params)
				if req.Notif {
					logger.Printf("jsonrpc2: --> notif: %s: %s\n", req.Method, params)
				} else {
					logger.Printf("jsonrpc2: --> request #%s: %s: %s\n", req.ID, req.Method, params)
				}

			case resp != nil:
				var method string
				if req != nil {
					method = req.Method
				} else {
					method = "(no matching request)"
				}
				switch {
				case resp.Result != nil:
					result, _ := json.Marshal(resp.Result)
					logger.Printf("jsonrpc2: --> result #%s: %s: %s\n", resp.ID, method, result)
				case resp.Error != nil:
					err, _ := json.Marshal(resp.Error)
					logger.Printf("jsonrpc2: --> error #%s: %s: %s\n", resp.ID, method, err)
				}
			}
		})(c)
		OnSend(func(req *Request, resp *Response) {
			switch {
			case req != nil:
				params, _ := json.Marshal(req.Params)
				if req.Notif {
					logger.Printf("jsonrpc2: <-- notif: %s: %s\n", req.Method, params)
				} else {
					logger.Printf("jsonrpc2: <-- request #%s: %s: %s\n", req.ID, req.Method, params)
				}

			case resp != nil:
				mu.Lock()
				method := reqMethods[resp.ID]
				delete(reqMethods, resp.ID)
				mu.Unlock()
				if method == "" {
					method = "(no previous request)"
				}

				if resp.Result != nil {
					result, _ := json.Marshal(resp.Result)
					logger.Printf("jsonrpc2: <-- result #%s: %s: %s\n", resp.ID, method, result)
				} else {
					err, _ := json.Marshal(resp.Error)
					logger.Printf("jsonrpc2: <-- error #%s: %s: %s\n", resp.ID, method, err)
				}
			}
		})(c)
	}
}
