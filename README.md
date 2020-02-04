# XRPC
A Simple RPC Framework

## Requirements
- 支持protobuf + struct + function三种方式注册服务
- 传输类型(tcp, kcp, ws, quic) + (tls, multiple stream)
- 上层中间件(调用对应的handler function之前执行)
- 底层插件(插件初始化/销毁， 连接建立/断开(如交换秘钥,协商压缩算法)+消息收/发后处理(加解密，解压缩))
- 基于protoc-gen-go, 可直接使用proto3文件
- 完整的单元测试+性能测试
- jaeger分布式链路追踪
- prometheus监控上报
- 服务发现，负载均衡，连接认证
- p2p
- 多语言客户端(本地tcp连client agent->server)，主要是要修改对应的proto-gen-xxx