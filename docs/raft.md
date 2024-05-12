# raft 操作相关API

## 集群操作API
### 1. 加入集群/从集群中移除

> 启动多个节点的时候，需要在leader 节点调用加入节点的请求，以便组成集群，节点数量推荐至少三个
> 如果一个节点down掉master节点会一直打印节点相关的日志， 可以新启动节点，也可以把老的节点启动
> 后重新加入。
>
> 如果要移除某个节点，则需要调用一下移除节点（最好在leader节点调用，
> 也可以在其他节点，但是不能在已经被移除的节点调用）， 移除的节点为 leader节点剩余节点则会自动选主，
> 移除的是其他节点，则leader节点不会再报错误信息，被移除节点后续也不会再参与选主。

POST `/api/raft/cluster`

```json
{
  "cmd": "remove",
  "nodeId": "1001",
  "addr": "127.0.0.1:9001"
}
```

返回：

```json
{
  "code": "0",
  "msg": "success"
}
```
> 注： cmd 可以是 remove 或者 join 分别对应从集群中移除，以及加入集群


### 2. 获取集群信息

GET   `/api/raft/info`

返回

```json
{
  "code": "0",
  "msg": "success",
  "data": {
    "leaderId": "1000",
    "leaderAddr": "127.0.0.1:9000",
    "peers": [
      {
        "id": "1000",
        "addr": "127.0.0.1:9000",
        "role": "leader",
        "state": "Voter"
      }
    ]
  }
}
```

## 二、 底层存储操作相关API

> 此类API在任意节点均可， 底层已经使用raft同步，非leader节点的写入请求会自动重定向到leader节点进行写入

### 1. 获取KEY

GET `/api/store/key?key=k11`

```json
{
  "code": "0",
  "msg": "success",
  "data": "kubters"
}
```

### 2. 设置修改/删除KEY

POST `/api/store/key`

```json
{
  "cmd": "set",
  "key": "k11",
  "value": "kubters",
  "exp": 1000000000
}
```

cmd 可以是 `set`, `del`, `setex`, 分别对应设置， 删除，以及设置带过期时间的key

> 返回值如下

```json
{
  "code": "0",
  "msg": "success"
}
```

### 3. 扫描以前缀开始的KEY

GET  `/api/store/scan?prefix=svc`

```json
{
  "code": "0",
  "msg": "success",
  "data": {
    "svc.demoService.pref.meta.10.10.10.10:8001": "{\"tags\":\"sh,hz\"}",
    "svc.demoService.pref.state.10.10.10.10:8001": "UP"
  }
}
```



## 附录：raft小笔记

1. 只有两台节点的情况下， 停掉master会导致选主失败，因为不符合过半原则, 推荐奇数台节点
2. 只有两台需要启动原来的节点，才能正常选主
3. leader节点会持续向其他节点给心跳，除非手动remove掉其他节点
4. 集群配置信息会始终存在每个节点，当其他节点down掉之后，会自动选主，down掉的节点启动的时候，会自动连入集群
5. remove 之后，再进行 addVoter 添加已经移除的节点不会成功， 需要重启（估计是remove的时候把raft监听也停掉了）
6. 被remove掉的节点，从中获取集群配置，是错误的， 也同步不到最新log

