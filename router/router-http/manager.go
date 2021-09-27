package router_http

import (
	"crypto/tls"
	"errors"
	"net"
	"strings"
	"sync"

	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/traffic"

	"github.com/valyala/fasthttp"

	"github.com/eolinker/eosc/listener"
)

var _ iManager = (*Manager)(nil)

var (
	errorCertificateNotExit = errors.New("not exist cert")
)

type iManager interface {
	Add(port int, id string, config *Config) error
	Del(port int, id string) error
	Cancel()
}

var manager = NewManager()

//Manager 路由管理器结构体
type Manager struct {
	locker    sync.Mutex
	routers   IRouters
	servers   map[int]*httpServer
	listeners map[int]net.Listener

	traffic traffic.ITraffic
}

type httpServer struct {
	tlsConfig *tls.Config
	port      int
	protocol  string
	srv       *fasthttp.Server
	certs     *Certs
}

//shutdown 关闭http服务器
func (h *httpServer) shutdown() {
	h.srv.Shutdown()
}

//GetCertificate 获取证书配置
func (h *httpServer) GetCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if h.certs == nil {
		return nil, errorCertificateNotExit
	}
	certificate, has := h.certs.Get(strings.ToLower(info.ServerName))
	if !has {
		return nil, errorCertificateNotExit
	}

	return certificate, nil
}

//Cancel 关闭路由管理器
func (m *Manager) Cancel() {
	m.locker.Lock()
	defer m.locker.Unlock()
	for p, s := range m.servers {
		s.shutdown()
		delete(m.servers, p)
	}

	for k, l := range m.listeners {
		l.Close()
		delete(m.listeners, k)
	}
}

//NewManager 创建路由管理器
func NewManager() *Manager {
	var traffic traffic.ITraffic
	bean.Autowired(&traffic)
	bean.AddInitializingBeanFunc(func() {

	})
	return &Manager{

		routers:   NewRouters(),
		servers:   make(map[int]*httpServer),
		listeners: make(map[int]net.Listener),
		locker:    sync.Mutex{},
	}
}

//Add 新增路由配置到路由管理器中
func (m *Manager) Add(port int, id string, config *Config) error {
	m.locker.Lock()
	defer m.locker.Unlock()

	router, isCreate, err := m.routers.Set(port, id, config)
	if err != nil {
		return err
	}
	if isCreate {
		s, has := m.servers[port]
		if !has {
			s = &httpServer{srv: &fasthttp.Server{}}

			s.srv.Handler = router.Handler()
			l, err := listener.ListenTcp("", port)

			if err != nil {
				return err
			}
			if config.Protocol == "https" {
				s.certs = newCerts(config.Cert)
				s.tlsConfig = &tls.Config{GetCertificate: s.GetCertificate}
				l = tls.NewListener(l, s.tlsConfig)
			}
			go s.srv.Serve(l)

			m.servers[port] = s
			m.listeners[port] = l
		}
	}
	return nil
}

//Del 将某个路由配置从路由管理器中删去
func (m *Manager) Del(port int, id string) error {
	m.locker.Lock()
	defer m.locker.Unlock()
	if r, has := m.routers.Del(port, id); has {
		//若目标端口的http服务器已无路由配置，则关闭服务器及listener
		if r.Count() == 0 {
			if s, has := m.servers[port]; has {
				err := s.srv.Shutdown()
				if err != nil {
					return err
				}
				delete(m.servers, port)
				m.listeners[port].Close()
				delete(m.listeners, port)
			}
		}
	}

	return nil

}

//Add 将路由配置加入到路由管理器
func Add(port int, id string, config *Config) error {
	return manager.Add(port, id, config)
}

//Del 将路由配置从路由管理器中删去
func Del(port int, id string) error {
	return manager.Del(port, id)
}
