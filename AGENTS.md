# jvmgo-playground 仓库协作规则

## 部署入口

- 本项目由 `deploy-webhook` 管理。平台仓库为 `git@github.com:LXL47/deploy-webhook.git`。
- 修改生产Dockerfile、Compose、健康检查、端口、域名、依赖或发布方式前，必须先读取平台仓库 `docs/architecture/deployment-overview.md` 和 `docs/operations/project-onboarding-checklist.md`。
- 本项目的发布声明是平台仓库 `config/projects/jvmgo-playground.cn.yaml`；具体端口、分支和发布模式以该声明为准，不在本文件复制动态值。
- 允许分支的普通Push会触发生产发布。提交和推送前必须按生产变更评估，不得把Git平台Ping或Gitee页面测试当成发布成功。
- 服务器现场、PostgreSQL策略和Git声明不一致时，先用 `deployctl plan/check` 判断漂移，禁止直接手改现场形成第二事实源。
