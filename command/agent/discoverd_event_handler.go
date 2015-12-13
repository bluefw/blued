package agent

import (
	"bytes"
	"github.com/bluefw/blued/discoverd"
	"github.com/bluefw/blued/discoverd/api"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/serf/serf"
	"log"
)

const (
	RSCommand  = "rs"
	URSCommand = "us"

	QRPCAddrCommand = "qr"
)

type DiscoverdEventHandler struct {
	discoverd *discoverd.Discoverd
	config    *Config
	logger    *log.Logger
}

func NewDiscoverdEventHandler(ds *discoverd.Discoverd, config *Config, logger *log.Logger) EventHandler {
	return &DiscoverdEventHandler{
		discoverd: ds,
		config:    config,
		logger:    logger,
	}
}
func (h *DiscoverdEventHandler) HandleEvent(e serf.Event) {
	var err error
	switch event := e.(type) {
	case serf.MemberEvent:
		if event.EventType() == serf.EventMemberFailed {
			err = h.onMemberFaild(event)
		} else if event.EventType() == serf.EventMemberLeave {
			err = h.onMemberLeave(event)
		}
	case serf.UserEvent:
		err = h.onUserEvent(event)
	case *serf.Query:
		err = h.onQuery(event)
	default:
		h.logger.Printf("[INFO] ds.event: Unknown event type: %s", e.EventType().String())
	}
	if err != nil {
		h.logger.Printf("[ERR] ds.event: Failed to handle event %v:%v", e, err)
	}
}

func (h *DiscoverdEventHandler) onMemberFaild(e serf.MemberEvent) error {
	h.logger.Printf("[INFO] ds.event: Handle member faild event.")
	for _, m := range e.Members {
		h.discoverd.RemoveRouterByHost(m.Name)
	}
	return nil
}

func (h *DiscoverdEventHandler) onMemberLeave(e serf.MemberEvent) error {
	h.logger.Printf("[INFO] ds.event: Handle member leave event.")
	for _, m := range e.Members {
		h.discoverd.RemoveRouterByHost(m.Name)
	}
	return nil
}

func (h *DiscoverdEventHandler) onUserEvent(e serf.UserEvent) error {
	switch e.Name {
	case RSCommand:
		var ias api.InnerAppService
		dec := codec.NewDecoder(bytes.NewReader(e.Payload), &codec.MsgpackHandle{})
		if err := dec.Decode(&ias); err != nil {
			return err
		}
		h.registerService(&ias)
	case URSCommand:
		var addr string
		dec := codec.NewDecoder(bytes.NewReader(e.Payload), &codec.MsgpackHandle{})
		if err := dec.Decode(&addr); err != nil {
			return err
		}
		h.unregisterService(addr)
	}
	return nil
}

func (h *DiscoverdEventHandler) onQuery(e *serf.Query) error {
	h.logger.Printf("[INFO] rpc:%s,e.Name:%s ", h.config.RPCAddr, e.Name)
	switch e.Name {
	case QRPCAddrCommand:
		h.logger.Printf("[INFO] rpc:%s ", h.config.RPCAddr)
		e.Respond([]byte(h.config.RPCAddr))
	}

	return nil
}

func (h *DiscoverdEventHandler) registerService(ias *api.InnerAppService) {
	h.discoverd.AddRouter(ias.NameAddr.Name, ias.NameAddr.Addr, ias.Services)
}
func (h *DiscoverdEventHandler) unregisterService(addr string) {
	h.discoverd.RemoveRouter(addr)
}
