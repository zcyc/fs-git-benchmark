# fs-git-benchmark

`fs-git-benchmark` 是一个高效的 Git 性能测试工具，旨在对 Git 操作（如 `clone`, `checkout`, `add`, `commit`, `push`）的运行速度进行基准测试。这个工具能够帮助您评估本地 Git 客户端与远程仓库之间不同操作的性能，也支持通过并发和多次重复执行进一步提高测试覆盖面。

## 功能说明

- 克隆指定的 Git 仓库到当前目录（使用单独的目标子目录）。
- 创建一个随机分支并切换到该分支。
- 创建一个随机文件，将其添加到版本控制中并提交。
- 将新分支推送到远程仓库。
- 支持多次执行整个操作，并支持并发运行以测试性能和资源利用。
- 对每个 Git 操作的时间进行记录，统计每个仓库的完整操作日志和耗时。

## 适用场景

- 测试性能：对常用 Git 操作的性能进行基准测试，发现瓶颈。
- 仓库维护：用于检查远程仓库与本地的交互时延和性能。
- 环境验证：验证是否因负载或网络延迟导致操作效率下降。

---

## 安装

### 使用源码安装

1. 确保已安装 [Golang](https://golang.org/dl/) 环境，版本在 `1.18` 或更高。
2. 安装：
   ```bash
   go install github.com/zcyc/fs-git-benchmark@latest
   ```
3. 运行程序：
   ```bash
   ./fs-git-benchmark
   ```
---

## 使用说明

### 参数说明

程序可以通过以下参数控制运行行为：

| 参数名称      | 类型    | 默认值 | 说明                                                      |
| ------------- | ------- | ------ | --------------------------------------------------------- |
| `-repo`       | string  | 无     | 需要操作的远程 Git 仓库地址（必需提供）。                 |
| `-concurrency`| integer | 1      | 并行测试的数量限制，表示同时运行多少个仓库操作流程。        |
| `-count`      | integer | 1      | 指定操作重复的次数，表示从远程仓库克隆的仓库实例数量。      |

### 使用示例

1. 测试单线程，执行 1 次：
   ```bash
   ./fs-git-benchmark -repo=https://github.com/example/repo.git -concurrency=1 -count=1
   ```

2. 并发测试：运行 4 个线程，分别执行 5 次：
   ```bash
   ./fs-git-benchmark -repo=https://github.com/example/repo.git -concurrency=4 -count=5
   ```

3. 查看帮助：
   ```bash
   ./fs-git-benchmark -h
   ```

输出示例：
```plaintext
2023/10/08 11:00:00 Starting git speed test with concurrency=4, count=2
2023/10/08 11:00:03 Logs for repo_0:
Cloning repository https://github.com/example/repo.git into ./git-speed-test-tmp/repo_0
Git command [clone https://github.com/example/repo.git ./git-speed-test-tmp/repo_0] executed successfully in 1.234567s
Checking out to new branch branch_0 in ./git-speed-test-tmp/repo_0
Git command [checkout -b branch_0] executed successfully in 0.123456s
...
2023/10/08 11:00:03 Logs for repo_1:
<similar logs for repo_1>
2023/10/08 11:00:04 Completed git speed test in 4.000000s
```

测试完成后，所有输出日志会自动记录每一步的操作并展示每一次 Git 操作的详细耗时。

---

## 工作原理

1. **克隆仓库**：
    - 程序会在当前工作目录下创建一个 `fs-git-benchmark-tmp/` 子目录，所有的克隆操作都会存放到这个目录中。
    - 每个仓库按照 `repo_0`、`repo_1` 等格式存放。

2. **并行操作**：
    - 使用 Go 的 Goroutines 并发执行多个仓库的操作，`-concurrency` 参数控制最大并发量。

3. **操作步骤**：
   每个仓库会运行以下操作：
    - 克隆 (`git clone`)
    - 创建随机分支并切换 (`git checkout -b <branch>`)
    - 创建随机文件并添加 (`git add`)
    - 提交 (`git commit -m "Test commit"`)
    - 推送 (`git push -u origin <branch>`)

4. **统计和日志**：
    - 每个仓库的整体操作都会被记录到单独的日志中，包括不同 Git 操作的耗时。
    - 所有仓库执行完成后，程序会打印每个仓库的完整日志。

5. **清理临时目录**：
   测试结束后，`fs-git-benchmark-tmp/` 子目录会自动删除，保持当前目录整洁。

---

## 注意事项

1. **远程仓库权限**：
    - 确保提供的仓库地址是有效的，并具有推送权限（推荐使用测试仓库）。

2. **Push 分支注意**：
    - 每次程序运行时都会推送多个随机分支到远程仓库，请确保仓库允许写入操作。
    - 使用完成后手动清理远程仓库中的无用分支。

3. **资源使用**：
    - 高 `-concurrency` 值可能会占用较多 CPU 和网络带宽，根据机器性能合理配置。

4. **调试日志**：
    - 所有操作日志会输出在标准输出中，便于调试。

---

## 示例场景

### 测试 Git 性能

使用高并发和多次执行的方式测试一个远程仓库在各种网络环境下的响应时间。例如：

```bash
./fs-git-benchmark -repo=https://github.com/example/repo.git -concurrency=10 -count=20
```

观察其中 `clone` 和 `push` 的耗时，可以确定网络延迟或仓库性能是否有瓶颈。

### 自动化测试仓库操作

在测试环境中，通过构建和提交新的分支及文件，用 `fs-git-benchmark` 验证脚本化的开发流程或环境部署的稳定性。

---

## 许可证

[MIT](./LICENSE)
