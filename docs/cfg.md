# 配置中心相关API

> 所有接了配置中心的服务都会与配置中心建立长连接，如果长连接断掉，则会新建一个长连接
>
> 应用在配置中心获取到配置后， 需要缓存到本地，然后与配置中心保持联系
>
> 当配置项被修改的时候，节点会向监听的应用发送被修改后的配置信息

## 管理相关API列表

TODO 由于内容比较多，暂不补充


## 配置推送格式 (Websocket)

```json
{
  "type": "cfg",
  "content": {
    "namespace": "default.yaml",
    "key": "config.item.users[0].name",
    "type": "change",
    "current": "Tom",
    "before": "Tom Lance"
  }
}
```

外层type 为固定值: `cfg`， 内层type 可以为如下三个值

1. `add` 表示配置新增
2. `remove` 表示配置删除
3. `change` 表示配置更改


## 应用连接所需要的API列表（需要进行验签）

POST `/api/cfg/client/app/{appId}`

请求参数：

```json
{
  "appId": "DemoService",
  "group": "default",
  "namespaces": "app.properties,cfg.yaml",
  "sharedNamespaces": "OrderService.common.shared.yaml,UserService.common.shared.json"
}
```

- 以上参数都为必填参数
- namespace 是指在指定 group下面所监听的 namespace 列表， 用英文逗号分割
- sharedNamespaces 是指一些共享的namespace, 必须是 appId.group.namespace 格式

## 附录， KEY PATTERN

```properties
# 当前版本 占位符分别为 appId, group, 与namespace
appConfigKeyPattern     = "app.cfg.current.%s.%s.%s"   
# 未发布的版本 占位符分别为 appId, group, 与namespace
appUnreleasedKeyPattern = "app.cfg.future.%s.%s.%s"
# 历史版本 占位符分别为 appId, group, namespace 和发布当时的时间戳
appHistoryKeyPattern    = "app.cfg.history.%s.%s.%s.%s"
```