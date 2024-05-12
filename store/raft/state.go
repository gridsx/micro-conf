package raft

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/raft"
)

type PeerState struct {
	Id    string `json:"id,omitempty"`
	Addr  string `json:"addr,omitempty"`
	Role  string `json:"role,omitempty"`
	State string `json:"state,omitempty"`
}

type ClusterState struct {
	LeaderId   string      `json:"leaderId,omitempty"`
	LeaderAddr string      `json:"leaderAddr,omitempty"`
	Peers      []PeerState `json:"peers,omitempty"`
}

func (c *ClusterState) String() string {
	d, _ := json.Marshal(c)
	return string(d)
}

func NewClusterInfo(rs *raft.Raft) *ClusterState {
	addr, id := rs.LeaderWithID()
	latestCfg := rs.GetConfiguration()
	cfg := latestCfg.Configuration()
	servers := cfg.Servers
	peers := make([]PeerState, 0, 3)
	for _, v := range servers {
		role := "follower"
		if strings.EqualFold(string(addr), string(v.Address)) {
			role = "leader"
		}
		peers = append(peers, PeerState{
			Id:    string(v.ID),
			Addr:  string(v.Address),
			Role:  role,
			State: v.Suffrage.String(),
		})
	}

	return &ClusterState{
		LeaderId:   string(id),
		LeaderAddr: string(addr),
		Peers:      peers,
	}
}
