package server

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/gurodrigues-dev/tcp-server-client/session"
)

type Request struct {
	ID     int         `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

type Response struct {
	ID     int         `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type TCPServer struct {
	listener       net.Listener
	sessions       map[string]*session.Session
	sessionsMu     sync.Mutex
	authManager    *AuthManager
	requestHandler *RequestHandler
}

func NewTCPServer(address string, requestHandler *RequestHandler, authManager *AuthManager) *TCPServer {
	return &TCPServer{
		sessions:       make(map[string]*session.Session),
		requestHandler: requestHandler,
		authManager:    authManager,
	}
}

func (s *TCPServer) Start(address string) error {
	var err error
	s.listener, err = net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("fail to start server %s: %w", address, err)
	}

	fmt.Printf("TCP server started and listening on %s\n", address)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("Fail to accept connection: %v\n", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *TCPServer) Stop() error {
	if s.listener != nil {
		s.sessionsMu.Lock()
		defer s.sessionsMu.Unlock()

		for key, session := range s.sessions {
			if err := session.Close(); err != nil {
				fmt.Printf("Fail to close session %s: %v\n", key, err)
			}
		}
		s.sessions = make(map[string]*session.Session)

		return s.listener.Close()
	}
	return nil
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.removeSession(conn.RemoteAddr().String())
		fmt.Printf("Connection closed: %s\n", conn.RemoteAddr().String())
	}()

	remoteAddr := conn.RemoteAddr().String()
	fmt.Printf("New connection established: %s\n", remoteAddr)

	session := session.NewSession("", conn)
	s.addSession(remoteAddr, session)

	for {
		var req Request
		err := s.readJSON(conn, &req)
		if err != nil {
			fmt.Printf("Fail to read json %s: %v\n", remoteAddr, err)
			return
		}

		fmt.Printf("Request received from %s: ID=%d, Method=%s\n", remoteAddr, req.ID, req.Method)

		var resp *Response
		if s.requestHandler != nil {
			resp, err = s.requestHandler.HandleRequest(session, &req)
			if err != nil {
				resp = &Response{
					ID:    req.ID,
					Error: err.Error(),
				}
			}
		} else {
			resp = &Response{
				ID:    req.ID,
				Error: "request handler not configured",
			}
		}

		if err := s.writeJSON(conn, resp); err != nil {
			fmt.Printf("fail to write JSON to %s: %v\n", remoteAddr, err)
			return
		}
	}
}

func (s *TCPServer) readJSON(conn net.Conn, v interface{}) error {
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("fail to decode JSON: %w", err)
	}
	return nil
}

func (s *TCPServer) writeJSON(conn net.Conn, v interface{}) error {
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(v); err != nil {
		return fmt.Errorf("fail to encode JSON: %w", err)
	}
	return nil
}

func (s *TCPServer) addSession(key string, session *session.Session) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	s.sessions[key] = session
}

func (s *TCPServer) removeSession(key string) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	delete(s.sessions, key)
}
