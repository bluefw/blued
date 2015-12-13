package discoverd

import (
	"encoding/base64"
	"github.com/bluefw/blued/discoverd/api"
	"github.com/bluefw/blued/discoverd/cluster"
	"github.com/bluefw/blued/discoverd/msd"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func TestRegMicroApp(t *testing.T) {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	cluster := cluster.NewMockCluster()
	repo := msd.NewDiscoverdRepo(logger, &cluster)
	hs := NewHttpServer(":8341", logger, repo)
	go hs.Start()

	ma := &api.MicroApp{
		Addr:      "http://a.com:8080/rs",
		Providers: []string{"a.b", "a.c"},
		Consumers: []string{"a.b", "a.c"},
	}

	makeRequest("PUT", "http://127.0.0.1:8341/msd/register", ma)
	addr := base64.StdEncoding.EncodeToString([]byte("http://a.com:8080/rs"))
	url := "http://127.0.0.1:8341/msd/router/" + addr
	r, err := makeRequest("GET", url, nil)
	if err != nil {
		t.Error("Error")
	}

	var rt api.RouterTable
	processResponseEntity(r, &rt, 200)

	assert.Equal(t, 2, len(rt.Routers))
	assert.Equal(t, ma.Addr, rt.Routers[0].Addrs[0])
	assert.Equal(t, ma.Addr, rt.Routers[1].Addrs[0])
}
