# Win 盘清理工具（WinDriveCleaner）

一个**零安装、绿色单文件**的 Windows C 盘垃圾清理工具。使用 Go 编译为静态可执行文件，
**无任何运行时依赖**，可在任意 Windows 机器上直接运行，兼容 **Windows 8 / 10 / 11**（32 位与 64 位）。

## 特性

- **两种模式**
  - **普通清理**：低风险、快速，清理系统/用户临时文件、Prefetch、缩略图缓存、DNS 缓存目录、Windows 更新下载缓存、WER 崩溃报告。
  - **深度清理**：在普通清理基础上额外处理回收站、系统级临时文件、WinSxS 旧组件、`Windows.old` 等，释放空间更多（需管理员权限）。
- **安全优先**
  - 绝不触碰 `System32` / `WinSxS` 受保护目录、`pagefile.sys`、`System Volume Information` 等危险目标。
  - 遇到被占用/无权限的文件自动跳过并计数，不中断整体流程。
  - 提供 `--dry-run` 演练模式：只统计可释放空间，**不删除任何文件**。
- **绿色便携**：单个 `.exe` 文件，拷到 U 盘即可在其他电脑使用，无需安装、无需 .NET 运行时。
- **中文界面 + 中文日志**，结果按分类列出并汇总释放空间。

## 使用方法

### 1. 交互菜单（推荐）
双击 `WinDriveCleaner.exe`，按提示选择：

```
================ Win 盘清理工具 ================
 系统: 系统盘=C:\  WINDIR=C:\Windows  架构=windows/amd64
 权限: 普通用户

  选择模式:
   1) 普通清理
   2) 深度清理(需管理员)
   3) 演练普通清理(不删除)
   4) 演练深度清理(不删除)
   q) 退出
  请输入:
```

### 2. 命令行

| 命令 | 说明 |
|------|------|
| `WinDriveCleaner.exe --standard` | 执行普通清理 |
| `WinDriveCleaner.exe --deep` | 执行深度清理（需要管理员；非管理员时自动触发 UAC 提权） |
| `WinDriveCleaner.exe --dry-run` | 演练：只统计，不删除 |
| `WinDriveCleaner.exe --no-uac` | 跳过 UAC 提权（以当前权限运行） |
| `WinDriveCleaner.exe --version` | 打印版本 |
| `WinDriveCleaner.exe --help` | 显示帮助 |

> 深度清理涉及系统级目录，程序在**非管理员**下运行会自动请求 UAC 提权；
> 提权成功后由新的管理员进程继续执行，并在结束时暂停等待回车，避免窗口一闪而过。

### 3. 进一步释放空间（可选）

深度清理完成后，如需再榨取空间，可手动在**管理员 PowerShell / CMD** 中运行：

```bat
DISM /Online /Cleanup-Image /StartComponentCleanup
```

这会清理 WinSxS 组件备份，通常可再释放数百 MB 到数 GB，但耗时较长。

## 从源码构建

需要 Go 1.21+（工具链零第三方依赖，纯标准库，可离线编译）：

```bat
go build -trimpath -ldflags "-s -w" -o WinDriveCleaner.exe .
```

交叉编译（在 Linux/macOS 上生成 Windows 版）：

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o WinDriveCleaner.exe .
```

生成的是**静态链接**的单文件 `.exe`，目标机器无需安装任何运行库。

## 安全说明

- 所有清理目标均按「临时/缓存/旧备份」语义白名单定义，未命中白名单的路径**不会被删除**。
- 被其他程序占用或无权限访问的文件会自动跳过，并在结果中计数，不会导致程序崩溃。
- 首次使用强烈建议先跑 `--dry-run` 查看可释放空间，再决定执行真实清理。

## 兼容性

| 项 | 支持 |
|----|------|
| 操作系统 | Windows 8 / 8.1 / 10 / 11 |
| 架构 | x86 / x64（本仓库默认产出 x64；x86 见 `构建`） |
| 依赖 | 无（静态单文件） |

生成 32 位版本：

```bat
go build -trimpath -ldflags "-s -w" -o WinDriveCleaner_386.exe .
```
（在 Windows 上配合 `GOARCH=386`，或交叉编译 `GOOS=windows GOARCH=386`。）

## 许可

MIT License。见 [LICENSE](LICENSE)。
