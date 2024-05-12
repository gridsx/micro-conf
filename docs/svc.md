# 服务治理相关API

## 设计介绍

每个服务分别具备分组，以及自定义meta的功能， 分组的功能是方便调用方指定具体分组而使用， 如果想更细粒度的控制流量，可以用meta元数据来控制，
通过客户端和服务端注册的meta， 可实现限流、tag路由、分区域等等功能， 

**key设计:**

1. 服务状态 `svc.state.{app}.{group}.{ip}:{port}`
2. meta信息 `svc.meta.{app}.{group}.{ip}:{port}`

## 客户端接口


### 3. 获取注册的服务的信息 （混用）

> 管理站可通过此接口获取用于展示
> 
> 客户端也可以通过此接口获取注册服务的信息，用于做自己的流量分发

POST `/api/svc/instances`

```json
{
  "app": "demoService",
  "group": "pref"
}
```

返回

```json
{
  "code": "0",
  "msg": "success",
  "data": [
    {
      "app": "demoService",
      "group": "pref",
      "ip": "10.10.10.10",
      "port": 8001,
      "state": "UP",
      "meta": {
        "tags": "sh,hz"
      }
    }
  ]
}
```
