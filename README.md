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
- **结果清晰、不闪退**
  - 每项清理后**列出具体路径**（`└ C:\...\Temp`），让你看清"到底清理了什么"。
  - 显示**清理前后磁盘剩余空间对比**与实测释放量。
  - 清理/演练结束后**默认停在"按回车键退出…"**，方便阅读；脚本自动化时用 `--no-pause` 自动结束。
- **登录自启（可选）**：交互菜单第 8) 项可注册登录自启，每次开机倒计时后按上次设置自动执行；倒计时内按任意键可取消，也可一键关闭下次自启。仅写当前用户 Run 键，不碰系统注册表、无需管理员。
- **软件缓存 / 大文件清单**：菜单 9) 查看 AppData 各软件缓存目录大小并勾选删除；菜单 10) 列出系统盘 ≥500MB 大文件 Top 20（只列不删）。
- **系统盘压缩**：菜单 11) 用 Windows 自带 `compact` LZX 就地压缩 `C:\Windows`，释放 1-3GB（Win10/11，可逆，需管理员）。
- **删除可追溯**：所有删除前会把「路径+大小」写入 `%LOCALAPPDATA%\WinDriveCleaner\CleanerLog_<日期>.txt`，随时可查"刚才到底删了什么"。

## 更新日志

- **v1.5.1**：新增「11) 系统盘压缩」。调用 Windows 10/11 自带 `compact /Compact /BaseFile:C:\compact.sys /EssentialDirectories`，把 `C:\Windows` 内文件就地 LZX 压缩，通常释放 **1-3GB**（机械硬盘收益最大）。可逆（`compact /decompress` 恢复）。自动检测系统是否支持（Win8/7 无 LZX 会跳过），需管理员。新增 `--compact` CLI 参数。
- **v1.5.0**：新增三项体验增强 + SHA256 校验。① 「9) AppData 软件缓存」：扫描 `%LOCALAPPDATA%`/`%APPDATA%` 下各软件目录大小 Top 15，可勾选删除指定软件目录（删除前把清单写入日志，可追溯）；② 「10) 大文件检测」：扫描系统盘 ≥500MB 文件 Top 20（排除系统目录/保留文件），只列不删，可按 B 序号打开所在文件夹；③ 删除前备份日志：所有删除（普通/深度/缓存/通讯）前把「路径+大小」追加写入 `%LOCALAPPDATA%\WinDriveCleaner\CleanerLog_<日期>.txt`，可追溯；④ GitHub release 附 `SHA256SUMS`，用户可 `certutil -hashfile` 或 `sha256sum -c` 验证下载未被篡改。
- **v1.4.0**：新增「登录自启 + 倒计时自动执行」。交互菜单加第 8) 项「自启动管理」：可一键开关登录自启、设置倒计时秒数、选普通/深度模式、关闭下次自启。登录自启经系统 Run 键（仅当前用户，无需管理员，不写系统注册表）以 `--auto` 启动，读取上次保存的设置自动执行；默认 30 秒倒计时，倒计时内按任意键可取消本次执行（取消不删任何文件），取消后提示如何用菜单 8 关闭自启。设置存于 `%LOCALAPPDATA%\WinDriveCleaner\settings.json`。
- **v1.3.1**：修复「视频/录屏检测」闪退 bug——`printMediaScan` 的 `top` 当候选文件少于默认 20 时未做上限收敛，导致 `files[i]` 数组越界 panic、窗口一闪而过；现改为 `top = min(top, len(files))`，任何文件数下都不越界。
- **v1.3.0**：① 清理过程**逐条实时输出**（`[1/7] … [7/7]` 边清边显示各项与释放量，不再等跑完才出结果）；② 交互菜单改为**循环**（做完一项自动返回菜单可继续做下一项，仅 `q` 退出；交互模式任务完成后不再暂停）；③ 修复两处路径静默失效（`Prefetch` 双斜杠、`Windows 更新下载缓存` 误拼成 `C:\Windows\Windows\…` 导致长期清 0 B）。
- **v1.2.0**：新增通讯软件（微信/QQ/企业微信）按 N 天过滤清理；新增视频/录屏检测（提示大小、可打开所在文件夹手动清理）；深度清理后自动把删不掉的残留临时文件压缩为 zip；新增 DISM 组件清理可选提示；新增 `--scan-comms` / `--scan-media` 只读扫描参数。
- **v1.1.0**：结果逐项列出清理路径；新增磁盘前后剩余空间对比；默认停在"按回车键退出"避免闪退；新增 `--no-pause` 参数。
- **v1.0.0**：首版，普通/深度两档清理 + UAC 提权 + dry-run 演练 + 双架构（x64/x86）。

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
