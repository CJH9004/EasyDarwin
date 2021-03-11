package rtsp

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

type Server struct {
	SessionLogger
	TCPListener *net.TCPListener
	TCPPort     int
	Stoped      bool
	pushers     map[string]*Pusher // Path <-> Pusher
	pushersLock sync.RWMutex
}

var Instance *Server = &Server{
	SessionLogger: SessionLogger{log.New(os.Stdout, "[RTSPServer]", log.LstdFlags|log.Lshortfile)},
	Stoped:        true,
	TCPPort:       554,
	pushers:       make(map[string]*Pusher),
}

func GetServer() *Server {
	return Instance
}

func (server *Server) Start() (err error) {
	var (
		logger   = server.logger
		addr     *net.TCPAddr
		listener *net.TCPListener
	)
	if addr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", server.TCPPort)); err != nil {
		return
	}
	if listener, err = net.ListenTCP("tcp", addr); err != nil {
		return
	}

	server.Stoped = false
	server.TCPListener = listener
	logger.Println("rtsp server start on", server.TCPPort)
	networkBuffer := 1048576
	for !server.Stoped {
		var (
			conn net.Conn
		)
		if conn, err = server.TCPListener.Accept(); err != nil {
			logger.Println(err)
			continue
		}
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			if err = tcpConn.SetReadBuffer(networkBuffer); err != nil {
				logger.Printf("rtsp server conn set read buffer error, %v", err)
			}
			if err = tcpConn.SetWriteBuffer(networkBuffer); err != nil {
				logger.Printf("rtsp server conn set write buffer error, %v", err)
			}
		}

		session := NewSession(server, conn)
		go session.Start()
	}
	return
}

func (server *Server) Stop() {
	logger := server.logger
	logger.Println("rtsp server stop on", server.TCPPort)
	server.Stoped = true
	if server.TCPListener != nil {
		server.TCPListener.Close()
		server.TCPListener = nil
	}
	server.pushersLock.Lock()
	server.pushers = make(map[string]*Pusher)
	server.pushersLock.Unlock()
}

func (server *Server) AddPusher(pusher *Pusher) bool {
	server.pushersLock.Lock()
	_, ok := server.pushers[pusher.Path()]
	if !ok {
		server.pushers[pusher.Path()] = pusher
		go pusher.Start()
	}
	server.pushersLock.Unlock()
	return true
}

func (server *Server) TryAttachToPusher(session *Session) (int, *Pusher) {
	server.pushersLock.Lock()
	attached := 0
	var pusher *Pusher = nil
	if _pusher, ok := server.pushers[session.Path]; ok {
		if _pusher.RebindSession(session) {
			session.logger.Printf("Attached to a pusher")
			attached = 1
			pusher = _pusher
		} else {
			attached = -1
		}
	}
	server.pushersLock.Unlock()
	return attached, pusher
}

func (server *Server) RemovePusher(pusher *Pusher) {
	logger := server.logger
	server.pushersLock.Lock()
	if _pusher, ok := server.pushers[pusher.Path()]; ok && pusher.ID() == _pusher.ID() {
		delete(server.pushers, pusher.Path())
		logger.Printf("%v end, now pusher size[%d]\n", pusher, len(server.pushers))
	}
	server.pushersLock.Unlock()
}

func (server *Server) GetPusher(path string) (pusher *Pusher) {
	server.pushersLock.RLock()
	pusher = server.pushers[path]
	server.pushersLock.RUnlock()
	return
}

func (server *Server) GetPushers() (pushers map[string]*Pusher) {
	pushers = make(map[string]*Pusher)
	server.pushersLock.RLock()
	for k, v := range server.pushers {
		pushers[k] = v
	}
	server.pushersLock.RUnlock()
	return
}

func (server *Server) GetPusherSize() (size int) {
	server.pushersLock.RLock()
	size = len(server.pushers)
	server.pushersLock.RUnlock()
	return
}
