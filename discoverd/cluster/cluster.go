package cluster

import (
	"bytes"
	"github.com/bluefw/blued/discoverd/api"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/serf/serf"
	"log"
)

const (
	RSCommand  = "rs"
	URSCommand = "us"
)

type Cluster interface {
	RegisterService(ss *api.AppService) error
	UnregisterService(addr string) error
}

func EncodeMessage(msg interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	encoder := codec.NewEncoder(buf, &codec.MsgpackHandle{})
	err := encoder.Encode(msg)
	return buf.Bytes(), err
}

func DecodeMessage(buf []byte, out interface{}) error {
	var handle codec.MsgpackHandle
	return codec.NewDecoder(bytes.NewReader(buf), &handle).Decode(&out)
}

type SerfCluster struct {
	serf   *serf.Serf
	name   string
	logger *log.Logger
}

func NewSerfCluster(serf *serf.Serf, logger *log.Logger) Cluster {
	return &SerfCluster{
		serf:   serf,
		name:   serf.LocalMember().Name,
		logger: logger,
	}
}

func (c *SerfCluster) GetSerf() *serf.Serf {
	return c.serf
}

func (c *SerfCluster) RegisterService(ss *api.AppService) error {
	ias := &api.InnerAppService{
		NameAddr: api.NameAddr{
			Name: c.name,
			Addr: ss.Addr,
		},
		Services: ss.Services,
	}
	payload, err := EncodeMessage(ias)
	if err != nil {
		return err
	}
	return c.serf.UserEvent(RSCommand, payload, true)
}

func (c *SerfCluster) UnregisterService(addr string) error {
	payload, err := EncodeMessage(addr)
	if err != nil {
		return err
	}
	return c.serf.UserEvent(URSCommand, payload, true)
}
