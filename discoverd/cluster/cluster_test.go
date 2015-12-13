package cluster

import (
	"blued/discoverd/api"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_encodeMessage(t *testing.T) {
	addr := "http://a.com:8080/rs"
	srvs := []string{"a.b.c.d", "a.b.c.e"}
	as := &api.AppService{
		Addr:     addr,
		Services: srvs,
	}

	raw, _ := EncodeMessage(as)
	var das api.AppService
	DecodeMessage(raw, &das)

	assert.Equal(t, addr, das.Addr)
	assert.Contains(t, srvs, das.Services[0])
	assert.Contains(t, srvs, das.Services[1])

}
