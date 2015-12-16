package msd

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/bluefw/blued/discoverd/api"
	"github.com/bluefw/blued/discoverd/cluster"
	"github.com/bluefw/blued/discoverd/util/cache"
	"log"
	"os"
	"sync"
	"time"
)

type DiscoverdRepo struct {
	apps    *cache.Cache
	ttl     time.Duration
	routers map[string]api.Router
	rtLock  sync.RWMutex

	cluster cluster.Cluster
	logger  *log.Logger
}

func NewDiscoverdRepo(cluster cluster.Cluster, ttl time.Duration, l *log.Logger) *DiscoverdRepo {
	if l == nil {
		l = log.New(os.Stderr, "", log.LstdFlags)
	}
	dr := &DiscoverdRepo{
		apps:    cache.NewCache(ttl, ttl),
		ttl:     ttl,
		routers: make(map[string]api.Router),
		cluster: cluster,
		logger:  l,
	}

	dr.apps.RegExpiredHandler(func(dm map[string]interface{}) {
		dr.OnAppExpired(dm)
	})
	return dr
}

func (s *DiscoverdRepo) OnAppExpired(dm map[string]interface{}) {
	s.logger.Printf("[INFO] msd: Expired app:%v", dm)
	s.rtLock.Lock()
	for k, _ := range dm {
		err := s.cluster.UnregisterService(k)
		if err != nil {
			s.logger.Printf("[ERR] msd.repo: Failed to send register event:%s", err)
		}
	}
	s.rtLock.Unlock()
}

func (s *DiscoverdRepo) Register(ma *api.MicroApp) {
	s.logger.Printf("[INFO] ds.msd: Registering app:%v", ma)
	s.apps.Set(ma.Addr, ma, cache.DefaultExpiration)

	err := s.cluster.RegisterService(&api.AppService{
		Addr:     ma.Addr,
		Services: ma.Providers,
	})
	if err != nil {
		s.logger.Printf("[ERR] msd.repo: Failed to send register event:%s", err)
	}
}

func (s *DiscoverdRepo) ListMicroApps() []api.MicroApp {
	ms := make([]api.MicroApp, s.apps.ItemCount())
	idx := 0
	for _, item := range s.apps.Items() {
		ma := item.Object.(*api.MicroApp)
		ms[idx] = *ma
		idx++
	}
	return ms
}

func (s *DiscoverdRepo) ListRouters() []api.Router {
	s.rtLock.RLock()
	defer s.rtLock.RUnlock()
	rs := make([]api.Router, len(s.routers))
	idx := 0
	for _, r := range s.routers {
		rs[idx] = r
		idx++
	}
	return rs
}

func (s *DiscoverdRepo) UpdateRouters(rs []api.Router) {
	s.logger.Printf("[INFO] ds.msd: Updating router table")
	s.rtLock.RLock()
	defer s.rtLock.RUnlock()

	for k := range s.routers {
		delete(s.routers, k)
	}
	for _, v := range rs {
		s.routers[v.Service] = v
	}
}

func (s *DiscoverdRepo) Refresh(addr string) *api.AppStatus {
	s.logger.Printf("[INFO] ds.msd: Refreshing app at:%s|", addr)
	isLive := s.apps.Refresh(addr, cache.DefaultExpiration)

	return &api.AppStatus{
		IsLive:   isLive,
		RouterCS: s.calcRouterCheckSum(addr),
	}
}

func (s *DiscoverdRepo) GetRouterTable(addr string) *api.RouterTable {
	s.rtLock.RLock()
	defer s.rtLock.RUnlock()
	return s.calcRouterTable(addr)
}

func (s *DiscoverdRepo) RemoveRouterByHost(node string) {
	s.logger.Printf("[INFO] ds.msd: Removing router by host:%s", node)
	s.rtLock.Lock()
	defer s.rtLock.Unlock()
	for k, v := range s.routers {
		addrs := v.Addrs
		size := len(addrs)
		for idx := size - 1; idx >= 0; idx-- {
			if addrs[idx].Node == node {
				addrs = append(addrs[:idx], addrs[idx+1:]...)
			}
		}

		if len(addrs) == 0 {
			delete(s.routers, k)
		} else {
			s.routers[k] = api.Router{
				Service:  k,
				Addrs:    addrs,
				Checksum: s.calcChecksum(addrs),
			}
		}
	}
}

func (s *DiscoverdRepo) RemoveRouter(addr string) {
	s.logger.Printf("[INFO] ds.msd: Removing router by addr:%s", addr)
	s.removeRouter(addr)
}

func (s *DiscoverdRepo) removeRouter(addr string) {
	for ms, router := range s.routers {
		addrs := router.Addrs
		size := len(addrs)
		for idx := size - 1; idx >= 0; idx-- {
			if addrs[idx].Addr == addr {
				addrs = append(addrs[:idx], addrs[idx+1:]...)
			}
		}

		if len(addrs) == 0 {
			delete(s.routers, ms)
		} else {
			s.routers[ms] = api.Router{
				Service:  ms,
				Addrs:    addrs,
				Checksum: s.calcChecksum(addrs),
			}
		}
	}
}

func (s *DiscoverdRepo) AddRouter(node string, addr string, mss []string) {
	s.logger.Printf("[INFO] ds.msd: Adding router:%s,%s{%v}", node, addr, mss)

	s.rtLock.Lock()
	defer s.rtLock.Unlock()

	// for shutdown micro app and upgrade very quickly.
	s.removeRouter(addr)
	for _, ms := range mss {
		router, exist := s.routers[ms]
		if !exist {
			nas := []api.NodeAddr{api.NodeAddr{Node: node, Addr: addr}}
			s.routers[ms] = api.Router{
				Service:  ms,
				Addrs:    nas,
				Checksum: s.calcChecksum(nas),
			}
		} else {
			addrs := router.Addrs
			var isExist bool
			for _, v := range addrs {
				if v.Addr == addr {
					isExist = true
					break
				}
			}
			if !isExist {
				addrs = append(addrs, api.NodeAddr{Node: node, Addr: addr})
				s.routers[ms] = api.Router{
					Service:  ms,
					Addrs:    addrs,
					Checksum: s.calcChecksum(addrs),
				}
			}
		}
	}
}

func (s *DiscoverdRepo) calcRouterCheckSum(addr string) string {
	ck := make([]byte, 16)
	ma, found := s.apps.Get(addr)
	if found {
		app := ma.(*api.MicroApp)
		for _, v := range app.Consumers {
			rt, exist := s.routers[v]
			if !exist {
				continue
			}
			for idx := 0; idx < 16; idx++ {
				ck[idx] += rt.Checksum[idx]
			}
		}
	}
	return hex.EncodeToString(ck)
}

func (s *DiscoverdRepo) calcRouterTable(addr string) *api.RouterTable {
	ma, found := s.apps.Get(addr)
	if !found {
		return nil
	}

	var routers []api.Router
	var checksum string

	app := ma.(*api.MicroApp)
	for _, v := range app.Consumers {
		router, exist := s.routers[v]
		if !exist {
			continue
		}
		routers = append(routers, router)
	}
	if len(routers) > 0 {
		ck := make([]byte, 16)
		for _, rt := range routers {
			for idx := 0; idx < 16; idx++ {
				ck[idx] += rt.Checksum[idx]
			}
		}

		checksum = hex.EncodeToString(ck)
	}

	return &api.RouterTable{
		Routers:  routers,
		Checksum: checksum,
	}
}

func (s *DiscoverdRepo) calcChecksum(ss []api.NodeAddr) []byte {
	hasher := md5.New()
	sum := make([]byte, 16)
	for _, s := range ss {
		hasher.Reset()
		hasher.Write([]byte(s.Addr))
		cs := hasher.Sum(nil)
		for idx := 0; idx < 16; idx++ {
			sum[idx] += cs[idx]
		}
	}
	return sum
}
