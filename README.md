# demo-api

多仓全栈演示的 Go API 基线项目。

当前版本提供一个最小的“同步生成报表”接口：收到请求后立即生成 CSV，并返回下载地址。后续可以把它改造成异步任务，用于演示 LoopForge 同时修改 Web、API、Worker 和 E2E 四个仓库。

## 运行

```bash
go run .
```

默认监听 `http://127.0.0.1:8080`。可通过以下环境变量调整：

- `PORT`：服务端口。
- `DEMO_DATA_DIR`：CSV 数据目录，默认 `.data`。

## 接口

创建同步报表：

```bash
curl -X POST http://127.0.0.1:8080/api/reports \
  -H 'Content-Type: application/json' \
  -d '{"name":"sales"}'
```

运行测试：

```bash
gofmt -w .
go vet ./...
go test ./...
```
