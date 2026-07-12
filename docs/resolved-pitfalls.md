# 已解决坑位

## 已跟踪文件软删除后无法整目录执行 git mv

- 问题现象：软删除目录内的已跟踪 EXE 后，`git mv jvmgo jvm/runtime` 报 `bad source`。
- 根因：`git mv` 会校验目录中所有索引条目，已移出工作树的文件仍存在于 Git 索引，因此源目录被视为不完整。
- 已验证可行解：保持索引不变，使用同盘普通目录移动迁移剩余文件，再由 Git 状态统一识别重命名和删除。
- 适用范围：需要按规定先软删除部分已跟踪文件、随后重组其父目录的 Git 工作树。
- 复发预防：目录重组前先检查待软删除文件；如果二者重叠，预先选择普通目录移动，并在提交前用 `git status` 和 `git diff --summary` 校验结果。

## Java 8 启动阶段依赖 Unsafe 模拟内存

- 问题现象：沙箱完全禁用 `sun.misc.Unsafe` 后，HelloWorld 在初始化 `System.out` 时失败。
- 根因：Java 8 的 `StreamEncoder` 会通过 NIO `ByteBuffer` 调用 `Unsafe.allocateMemory`，该路径发生在用户 `main` 方法之前。
- 已验证可行解：保留 JVM 原有的纯 Go 模拟地址空间，将申请大小纳入受管分配累计预算，并继续依赖容器内存硬限制。
- 适用范围：使用 Java 8 `rt.jar` 且需要初始化标准输出的本项目生产 JVM。
- 复发预防：修改 native 白名单后必须运行真实 Java 8 HelloWorld，不能只做 Go 包编译检查。

## Windows javac 诊断输出乱码

- 问题现象：Windows JDK 8 返回的编译错误经过 JSON 接口后出现替换字符。
- 根因：`javac` 默认使用系统语言和本地代码页输出诊断，而 API 响应统一声明 UTF-8。
- 已验证可行解：固定传入 `-J-Duser.language=en -J-Duser.country=US -J-Dfile.encoding=UTF-8`，错误输出稳定为 UTF-8 英文。
- 适用范围：跨 Windows 和 Linux 容器运行 JDK 8 编译器的服务。
- 复发预防：集成测试必须包含编译失败用例，并校验响应是有效 UTF-8。
