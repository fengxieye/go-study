package xclient

import (
	"log"
	"net/http"
	"strings"
	"time"
)

/*
主要是更新可用的服务。
向registry发送请求得到服务列表。
选择所用的服务还是按MultiServersDiscovery来
*/

type RegistryDiscovery struct {
	*MultiServersDiscovery
	registry   string        //中心服务ip
	timeout    time.Duration //更新间隔
	lastUpdate time.Time     // 上次更新时间
}

const defaultUpdateTimeout = time.Second * 10

func NewRegistryDiscovery(registerAddr string, timeout time.Duration) *RegistryDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	d := &RegistryDiscovery{
		MultiServersDiscovery: NewMultiServerDiscovery(make([]string, 0)),
		registry:              registerAddr,
		timeout:               timeout,
	}

	return d
}

func (d *RegistryDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

func (d *RegistryDiscovery) Refresh() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}

	//去注册中心请求得到所有服务
	log.Println("RegistryDiscovery.Refresh: refresh servers from resigtry,", d.registry)
	resp, err := http.Get(d.registry)
	if err != nil {
		log.Println("RegistryDiscovery.Refresh get err:", err)
		return err
	}

	//划分
	servers := strings.Split(resp.Header.Get("X-Grpc-Servers"), ",")
	d.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		if strings.TrimSpace(server) != "" {
			d.servers = append(d.servers, strings.TrimSpace(server))
		}
	}

	d.lastUpdate = time.Now()
	return nil
}

func (d *RegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}

	return d.MultiServersDiscovery.Get(mode)
}

func (d *RegistryDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}

	return d.MultiServersDiscovery.GetAll()
}
