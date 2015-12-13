package msd

import (
	"blued/discoverd/api"
	"encoding/hex"
	"testing"
	"time"
)

func Test_Register(t *testing.T) {
	sr := createDiscoverdRepo(nil)
	mss := []string{"a.b", "a.c"}
	url := "http://a.com:8080/rs"
	oma := &api.MicroApp{Addr: url, Providers: mss}
	sr.Register(oma)
	v, _ := sr.apps.Get(url)
	nma := v.(*api.MicroApp)
	if oma.Addr != nma.Addr {
		t.Error("Addr is not equal")
	}
	if len(oma.Providers) != len(nma.Providers) {
		t.Error("length of mss is not equal")
	}

	for idx := 0; idx < len(oma.Providers); idx++ {
		if oma.Providers[idx] != nma.Providers[idx] {
			t.Errorf("mss[%d]=%s, expect:%s", idx, oma.Providers[idx], nma.Providers[idx])
		}
	}

	if sr.routers["a.b"].Addrs[0] != url || sr.routers["a.c"].Addrs[0] != url {
		t.Error("service is not register to consumer")
	}
}

func Test_TTL(t *testing.T) {
	sr := createDiscoverdRepo(nil)
	mss := []string{"a.b", "a.c"}
	url := "http://a.com:8080/rc"
	si := &api.MicroApp{Addr: url, Providers: mss}
	sr.Register(si)
	time.Sleep(1200 * time.Millisecond)
	_, found := sr.apps.Get(url)
	if found {
		t.Errorf("si is not expirated with %d second", 1)
	}

	if _, exist := sr.routers["a.b"]; exist {
		t.Logf("cr=%v", sr.routers)
		t.Errorf("app is not expirated in cr with %d second", 1)
	}

	sr.Register(si)
	time.Sleep(800 * time.Millisecond)
	sr.Refresh(url)
	time.Sleep(800 * time.Millisecond)
	_, found = sr.apps.Get(url)
	if !found {
		t.Errorf("si is expirated with %d millisecond", 800)
	}
}

func Test_removeRouter(t *testing.T) {
	sr := createDiscoverdRepo(nil)
	url := "http://a.com:8080/rc"
	sr.routers["a.b"] = api.Router{Service: "a.b", Addrs: []string{url}}
	sr.routers["a.c"] = api.Router{Service: "a.c", Addrs: []string{url}}

	sr.removeRouter(url, []string{"a.b", "a.c"})
	if _, exist := sr.routers["a.b"]; exist {
		t.Errorf("app is not removed in router %v", sr.routers["a.b"])
	}
}

func Test_OnAppExpired(t *testing.T) {
	sr := createDiscoverdRepo(nil)
	url := "http://a.com:8080/rc"
	sr.routers["a.b"] = api.Router{Service: "a.b", Addrs: []string{url}}
	sr.routers["a.c"] = api.Router{Service: "a.c", Addrs: []string{url}}

	dm := make(map[string]interface{})
	dm[url] = &api.MicroApp{Addr: url, Providers: []string{"a.b", "a.c"}}
	sr.OnAppExpired(dm)

	if _, exist := sr.routers["a.b"]; exist {
		t.Errorf("app is not removed in router %v", sr.routers["a.b"])
	}
}

func Test_CalcSign(t *testing.T) {
	sr := createDiscoverdRepo(nil)
	s1 := hex.EncodeToString(sr.calcChecksum([]string{"a.b", "a.c"}))
	s2 := hex.EncodeToString(sr.calcChecksum([]string{"a.c", "a.b"}))
	if s1 != s2 {
		t.Errorf("sign s1[%s] != s2[%s]", s1, s2)
	}
}
