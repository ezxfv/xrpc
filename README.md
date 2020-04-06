# XRPC
A Simple RPC Framework

## Requirements
- 支持proto3(定制proto-gen-go)和go interface(基于go ast解析)生成桩代码，或者直接使用函数地址Call(reflect实现)，入参和返回值都是[]byte
- 自定义协议: (tcp, kcp) x (tls, multiple stream)
- 服务注册(consul/chord dht)
- 插件系统
  - jaeger分布式链路追踪
  - prometheus监控上报
  - 特定日志
  - 连接黑白名单
  - 连接认证
  - 加密数据
  - 服务注册
  - 限流

## TODO
- Transport
  - gRPC http2
  - quic http3
- 适配MOSN私有协议XProtocol和HTTP2，quic的适配需要实现udp filter，暂时不行
- 连接超时优雅关闭，自动重连, 连接池