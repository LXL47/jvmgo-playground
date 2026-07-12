# JVMGO Playground

JVMGO Playground 是一个 Go 单仓库项目：保留《自己动手写 Java 虚拟机》的章节代码，同时提供在线 Java 8 代码编辑、编译和执行页面。后端与 Runner 均使用 Go，部署入口统一为 Docker Compose。

公开仓库：<https://github.com/LXL47/jvmgo-playground>

`main` 分支通过 GitHub Actions 执行 Go 测试、静态检查和两个 Docker 镜像的构建验证，不包含自动部署。

## 架构

```text
浏览器
  -> API（静态页面、限流、请求校验）
  -> Runner（独立任务目录、javac、超时与产物校验）
  -> jvmgo 子进程（指令、堆、数组、输出预算）
```

API 暴露 `8080` 端口。Runner 只连接 Compose 内部网络，不映射宿主机端口，也不挂载宿主机工作目录或 Docker Socket。

## 目录

```text
apps/                 Go API、Runner 和内嵌前端
config/               提交到 Git 的 YAML 配置
examples/java/        Java 示例程序
jvm/chapters/         ch01 到 ch11 教学源码
jvm/runtime/ch11/     用于在线执行的生产版 JVM
compose.yaml          默认加固部署
compose.gvisor.yaml   可选 gVisor 运行时覆盖
```

## 启动

需要 Docker Engine 和 Docker Compose v2：

```bash
docker compose up --build -d
docker compose ps
```

访问 <http://localhost:8080>。停止服务：

```bash
docker compose down
```

服务器已经安装并配置 `runsc` 时，推荐让 Runner 运行在 gVisor 中：

```bash
docker compose -f compose.yaml -f compose.gvisor.yaml up --build -d
```

生产服务器统一使用`compose.production.yaml`。该配置只把API发布到宿主机回环地址`127.0.0.1:8001`，Runner不发布端口并强制使用gVisor：

```bash
docker compose -f compose.production.yaml config
docker compose -f compose.production.yaml up --build -d
curl -fsS http://127.0.0.1:8001/healthz
```

生产配置要求Docker已经登记`runsc`运行时；缺少gVisor时应停止部署，不得回退为公开匿名代码直接运行在默认runc中。

Docker构建阶段固定使用国内Go模块代理，避免国内服务器直连`proxy.golang.org`导致生产构建超时；模块校验仍通过`sum.golang.google.cn`执行。

## 配置

项目不使用 `.env`。API 配置位于 `config/api.yaml`，Runner 和 JVM 沙箱预算位于 `config/runner.yaml`。

默认执行预算：

| 预算 | 默认值 | 作用 |
| --- | ---: | --- |
| 字节码指令 | 2,000,000 | 阻止无限循环长期占用 CPU |
| 受管分配累计量 | 32 MiB | 限制对象、数组和字符串等累计分配 |
| 单个数组长度 | 100,000 | 阻止超大数组及多维数组放大 |
| 标准输出 | 32 KiB | 达到上限后主动终止任务 |
| 执行墙钟 | 2 秒 | JVM 自身卡死时由 Runner 强制终止进程组 |

这些值面向教学示例，不适合计算密集型程序。调整时应先测量正常样例，再按峰值的 3 到 5 倍设置，不应只放大某一个维度。

## API

执行接口不兼容旧版 `/jvmgo/go`。

```http
POST /api/v1/executions
Content-Type: application/json

{"source":"public class Main { public static void main(String[] args) {} }"}
```

响应状态包括 `success`、`compile_error`、`runtime_error`、`sandbox_limit`、`timeout`、`output_limit` 和 `busy`。`GET /api/v1/runtime` 返回当前公开预算，`GET /healthz` 用于容器健康检查。

## 本地开发

```bash
go test ./apps/... ./jvm/runtime/...
go build ./jvm/runtime/ch11
```

`go.work` 已连接 `apps` 与 `jvm/runtime` 两个模块。Windows 已提供 `config/api.windows.yaml` 和 `config/runner.windows.yaml`，其中 JDK 8 安装路径需要与本机一致。

## 安全边界

- 用户源码不会经过 shell，`javac` 固定启用 `-proc:none`。
- 每个请求使用独立临时目录，完成后清理。
- 生产 JVM 运行在独立子进程中，panic 不会带崩 API。
- 沙箱模式拒绝文件元数据探测；`sun.misc.Unsafe` 使用纯 Go 模拟内存并计入受管分配预算，不接触宿主机原始地址。
- Compose 启用非 root、只读根文件系统、能力清空、PID/CPU/内存限制和内部网络。
- 页面同源访问 API，不开放任意 CORS，并设置 CSP 等安全响应头。

普通 Docker 不是面向敌对代码的绝对安全边界。长期匿名公开服务应启用 gVisor，并在入口层增加独立的限流、封禁和监控。

## 兼容范围

生产运行时基于 Java 8 类库结构，只实现项目已有的 JVM 指令和 native 方法集合。它用于学习和演示，不等同于 OpenJDK，也不保证运行任意 Java 程序。
