package raft

import (
	"github.com/gridsx/micro-conf/config"
	"github.com/hashicorp/raft"
	"github.com/winjeg/go-commons/cryptos"
	"github.com/winjeg/go-commons/http"

	"encoding/json"
	"errors"
	"fmt"
	http2 "net/http"
	"strconv"
	"strings"
)

const (
	keyPath          = "/api/store/key"    // set key
	removeServerPath = "/api/raft/cluster" // remove server from cluster
)

var cfg = config.App
var token = cryptos.Sha1([]byte(cfg.JWT.Secret))

// RedirectKeyRequest redirect the request to leader node
func RedirectKeyRequest(rs *raft.Raft, cmd, k, v string, exp int64) error {
	clusterInfo := NewClusterInfo(rs)
	if clusterInfo == nil {
		return errors.New("wrong cluster info")
	}
	addr, err := getLeaderAddr(clusterInfo.LeaderAddr)
	if err != nil {
		return err
	}
	contentMap := map[string]interface{}{"cmd": cmd, "key": k, "value": v, "exp": exp}
	return requestRemote(addr, keyPath, contentMap)
}

// RedirectRaftRequest redirect raft operation to leader node
func RedirectRaftRequest(rs *raft.Raft, nodeId, addr string) error {
	clusterInfo := NewClusterInfo(rs)
	if clusterInfo == nil {
		return errors.New("wrong cluster info")
	}
	leaderAddr, err := getLeaderAddr(clusterInfo.LeaderAddr)
	if err != nil {
		return err
	}
	contentMap := map[string]interface{}{"cmd": "remove", "addr": addr, "nodeId": nodeId}
	return requestRemote(leaderAddr, removeServerPath, contentMap)
}

func requestRemote(addr, path string, contentMap map[string]interface{}) error {
	d, _ := json.Marshal(contentMap)
	respStr, err := http.DoRequest("POST", addr+path, string(d), http2.Header{"_inner_auth": []string{token}})
	if err != nil {
		return err
	}
	respMap := map[string]interface{}{}
	jsonErr := json.Unmarshal([]byte(respStr), &respMap)
	if jsonErr != nil {
		return err
	}
	if v, ok := respMap["code"]; ok {
		if v != nil {
			if code, ok := v.(string); ok && strings.EqualFold(code, "0") {
				return nil
			}
		}
	}
	return errors.New("redirect to leader request error")
}

func getLeaderAddr(addr string) (string, error) {
	if len(addr) == 0 {
		return "", errors.New("wrong leader addr")
	}
	arr := strings.Split(addr, ":")
	if len(arr) != 2 {
		return "", errors.New("wrong leader addr")
	}
	p, err := strconv.Atoi(arr[1])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s:%d", arr[0], p-1000), nil
}
