package cfg

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gridsx/micro-conf/service/app"
	"github.com/gridsx/micro-conf/service/base"
	"github.com/gridsx/micro-conf/service/conn"
	"github.com/gridsx/micro-conf/store/raft"
	jsoniter "github.com/json-iterator/go"
	"github.com/kataras/iris/v12"
	"github.com/winjeg/go-commons/log"
	"github.com/winjeg/irisword/ret"
)

const (
	waitClientReconnectTime = 30 // 等待客户端重连的时间
	maxRetryCount           = 20 // 最大重试次数
)

type ConfigChangeRequest struct {
	AppId string         `json:"appId,omitempty"`
	Group string         `json:"group,omitempty"`
	Diff  *NamespaceDiff `json:"diff,omitempty"`
}

// AcceptConfigChange  只允许集群内节点之间相互调用
func acceptConfigChange(ctx iris.Context) {
	req := new(ConfigChangeRequest)
	if err := ctx.ReadJSON(req); err != nil {
		ret.BadRequest(ctx)
		return
	}
	doPush(req)
	ret.Ok(ctx)
}

// 根据namespace 获取监听的ip列表
// 针对每个监听的ip列表进行push
// 如果自己不是Leader的话， 那么需要把消息打包发给其他节点，等待其他节点ack
func pushChange(appId, group string, diff *NamespaceDiff) {
	request := &ConfigChangeRequest{
		AppId: appId,
		Group: group,
		Diff:  diff,
	}
	// 先把连接到自己这边的push一遍
	doPush(request)
	// 然后把消息存起来， 发给每个节点
	info := base.GetClusterInfo()
	if info == nil || len(info.Peers) == 0 {
		log.GetLogger(nil).Errorln("pushChange - error getting cluster info")
	} else {
		// 针对每个节点，进行变化转发， 转发后， 接收到消息的节点，找出自己本地；连接的应用，并把变化推送出去
		for _, p := range info.Peers {
			if base.IsSelf() {
				continue
			}
			go func() {
				if sendFollowerChange(p, request) {
					return
				}
				for i := 0; i <= maxRetryCount; i++ {
					time.Sleep(time.Second * time.Duration(i))
					if sendFollowerChange(p, request) {
						break
					}
				}
			}()
		}
	}
}

// 发送成功则不存储，发送失败，存储下来，用于后续重发
func sendFollowerChange(p raft.PeerState, request *ConfigChangeRequest) bool {
	arr := strings.Split(p.Addr, ":")
	raftPort, _ := strconv.Atoi(arr[1])
	data, err := jsoniter.Marshal(request)
	if err != nil {
		log.GetLogger(nil).Errorln("sendFollowerChange - json err: " + err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/api/cfg/listen", arr[0], raftPort-1000)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.GetLogger(nil).Errorln("sendFollowerChange - http err: " + err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		log.GetLogger(nil).Errorf("sendFollowerChange - http err, code : %d\n", resp.StatusCode)
	}
	respData, err := io.ReadAll(resp.Body)
	responseData := new(ret.Ret)
	if err := jsoniter.Unmarshal(respData, responseData); err != nil {
		log.GetLogger(nil).Errorf("sendFollowerChange - json unmarshal err: %s\n", err.Error())
	}
	if responseData.Code == "0" {
		return true
	}
	return false
}

func doPush(request *ConfigChangeRequest) {
	appId, group, diff := request.AppId, request.Group, request.Diff
	instances := app.GetNamespaceInstances(appId, group, diff.Namespace)
	for _, inst := range instances {
		wsKey := fmt.Sprintf(app.ClientKeyFormat, appId, inst.IP, inst.Port)
		client := conn.GetClient(wsKey)
		// 如果client断开，30秒内重新连上是没关系的
		if client == nil {
			for i := 0; i < waitClientReconnectTime; i++ {
				time.Sleep(time.Second)
				client = conn.GetClient(wsKey)
				if client != nil {
					break
				}
			}
		}

		if client != nil {
			if len(diff.Added) > 0 {
				for k, v := range diff.Added {
					event := &ConfigChangeEvent{
						Namespace: diff.Namespace,
						Key:       k,
						Type:      ConfigAdd,
						Current:   v,
					}
					appEvent := app.AppEvent{
						Type:    app.ConfigChange,
						Content: event,
					}
					client.Send(appEvent.String())
				}
			}
			if len(diff.Removed) > 0 {
				for k, v := range diff.Removed {
					event := &ConfigChangeEvent{
						Namespace: diff.Namespace,
						Key:       k,
						Type:      ConfigRemove,
						Current:   v,
					}
					appEvent := app.AppEvent{
						Type:    app.ConfigChange,
						Content: event,
					}
					client.Send(appEvent.String())
				}
			}

			if len(diff.Changed) > 0 {
				for k, v := range diff.Changed {
					event := &ConfigChangeEvent{
						Namespace: diff.Namespace,
						Key:       k,
						Type:      ConfigChange,
						Current:   v.Right,
						Before:    v.Left,
					}
					appEvent := app.AppEvent{
						Type:    app.ConfigChange,
						Content: event,
					}
					client.Send(appEvent.String())
				}
			}
		}
	}
}
