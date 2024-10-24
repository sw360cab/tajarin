package json

import (
	"encoding/json"
	"fmt"
)

const (
	ConfigPersistentPeers = "persistent_peers"
	DefaultListenAddress  = "0.0.0.0:8546"
	DefaultP2PPort        = "26656"
)

type JsonTajarinRequest struct {
	Name      string `json:"name" validate:"required"`
	Address   string `json:"address" validate:"required"`
	PubKey    string `json:"pubKey" validate:"required"`
	P2PNodeId string `json:"p2pNode" validate:"required"`
	P2PHost   string `json:"p2pHost" validate:"required"`
	P2PPort   string `json:"p2pPort,omitempty"`
}

func (jt *JsonTajarinRequest) ToP2PEndpointString() string {
	return fmt.Sprintf("%s@%s:%s", jt.P2PNodeId, jt.P2PHost, jt.P2PPort)
}

type JsonTajarinResponse struct {
	Genesis json.RawMessage   `json:"genesis,omitempty"`
	Config  map[string]string `json:"config,omitempty"`
	Error   string            `json:"error,omitempty"`
}
