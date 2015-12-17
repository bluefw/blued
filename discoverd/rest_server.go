package discoverd

import (
	"github.com/bluefw/blued/discoverd/msd"
	"github.com/braintree/manners"
	"github.com/gin-gonic/gin"
	"log"
)

func StartRestServer(addr string, repo *msd.DiscoverdRepo, logger *log.Logger, shutdownCh chan struct{}) {
	rs := msd.NewServiceResource(repo, logger)

	router := gin.Default()
	router.PUT("/msd/register", func(c *gin.Context) {
		rs.RegMicroApp(c)
	})

	router.GET("/msd/fresh/:addr", func(c *gin.Context) {
		rs.Refresh(c)
	})

	router.GET("/msd/fetch/:addr", func(c *gin.Context) {
		rs.GetRouterTable(c)
	})

	go func() {
		err := manners.ListenAndServe(addr, router)
		close(shutdownCh)
		logger.Fatalf("[ERR] rest server stopped: %v", err)
	}()
}

func ShutdownRestServer() bool {
	return manners.Close()
}
