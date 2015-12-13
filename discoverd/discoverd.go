package discoverd

import (
	"github.com/bluefw/blued/discoverd/api"
	"github.com/bluefw/blued/discoverd/cluster"
	"github.com/bluefw/blued/discoverd/msd"

	"github.com/hashicorp/serf/serf"
	"io"
	"log"
	"time"
)

type Context struct {
	RestAddr   string
	Serf       *serf.Serf
	ServiceTTL int
	Logger     *log.Logger
	ShutdownCh chan struct{}
}

type Discoverd struct {
	repo       *msd.DiscoverdRepo
	logger     *log.Logger
	shutdownCh chan struct{}
}

func Create(restAddr string, serviceTTL int, serf *serf.Serf, logOutput io.Writer) *Discoverd {
	logger := log.New(logOutput, "", log.LstdFlags)
	ttl := time.Duration(serviceTTL) * time.Second
	cluster := cluster.NewSerfCluster(serf, logger)
	repo := msd.NewDiscoverdRepo(cluster, ttl, logger)
	shutdownCh := make(chan struct{})
	StartRestServer(restAddr, repo, logger, shutdownCh)

	return &Discoverd{
		repo:       repo,
		logger:     logger,
		shutdownCh: shutdownCh,
	}
}

func (d *Discoverd) Shutdown() {
	d.logger.Println("[INFO] discoverd: shutting down ...")
	ShutdownRestServer()
}

// ShutdownCh returns a channel that can be used to wait for
// discoverd to shutdown.
func (d *Discoverd) ShutdownCh() <-chan struct{} {
	return d.shutdownCh
}

func (s *Discoverd) ListMicroApps() []api.MicroApp {
	return s.repo.ListMicroApps()
}

func (s *Discoverd) ListRouters() []api.Router {
	return s.repo.ListRouters()
}

func (s *Discoverd) UpdateRouters(rs []api.Router) {
	s.repo.UpdateRouters(rs)
}

func (s *Discoverd) AddRouter(name string, addr string, mss []string) {
	s.repo.AddRouter(name, addr, mss)
}

func (s *Discoverd) RemoveRouter(addr string) {
	s.repo.RemoveRouter(addr)
}

func (s *Discoverd) RemoveRouterByHost(name string) {
	s.repo.RemoveRouterByHost(name)
}
