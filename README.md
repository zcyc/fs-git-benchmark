# git-bench

`git-bench` is an efficient Git performance benchmarking tool designed to measure the execution speed of Git operations, such as `clone`, `checkout`, `add`, `commit`, and `push`. This tool helps evaluate performance between local Git clients and remote repositories. It also supports concurrent and repeated executions to broaden testing coverage.

## Features

- Clone a specified Git repository into the current directory (using separate target subdirectories).
- Create a random branch and switch to it.
- Create a random file, add it to version control, and commit the changes.
- Push the new branch to the remote repository.
- Supports multiple executions of the entire workflow and concurrent operations to test performance and resource usage.
- Records execution time for each Git operation, providing detailed logs and timing statistics for every repository.

## Use Cases

- **Performance Testing**: Benchmark commonly used Git operations to identify bottlenecks.
- **Repository Maintenance**: Examine latency and performance between remote repositories and local setups.
- **Environment Validation**: Verify whether high loads or network delays impact operational efficiency.

---

## Installation

### Install from Source

1. Ensure [Golang](https://golang.org/dl/) is installed with version `1.18` or higher.
2. Install the tool:
   ```bash
   go install github.com/zcyc/git-bench@latest
   ```

3. Run the tool:
   ```bash
   ./git-bench
   ```

---

## Usage Instructions

### Parameters

You can customize the behavior of this tool with the following parameters:

| Parameter Name    | Type    | Default | Description                                                      |
|-------------------|---------|---------|------------------------------------------------------------------|
| `-repo`           | string  | None    | The URL of the remote Git repository to operate on (required).   |
| `-concurrency`    | integer | 1       | The number of concurrent workflows to run.                       |
| `-count`          | integer | 1       | Number of times to repeat the operation.                         |
| `-use-go-git`     | boolean | false   | Whether to use `go-git`. By default, the tool uses the local Git client. |
| `-ssh-key`        | string  | None    | Path to the SSH key (absolute path) for `go-git` to use.         |

### Examples

1. Single-threaded test using the local Git client, run 1 time:
   ```bash
   ./git-bench -repo=https://github.com/example/repo.git -concurrency=1 -count=1
   ```

2. Concurrent test using `go-git`: Run with 4 threads, 5 executions each:
   ```bash
   ./git-bench -repo=https://github.com/example/repo.git -concurrency=4 -count=5 -use-go-git=true -ssh-key=/Users/charlie/.ssh/id_rsa
   ```

3. Display help:
   ```bash
   ./git-bench -h
   ```

Sample Output:
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

After the test, all output logs record the details of each step, including the time taken for each Git operation.

---

## How It Works

1. **Clone Repository**:
    - The program creates a `git-bench-tmp/` subdirectory in the current working directory. All cloned repositories are stored in this directory.
    - Each repository is stored under a directory named `repo_0`, `repo_1`, etc.

2. **Concurrent Execution**:
    - The tool uses Go's Goroutines to execute multiple workflows concurrently. The `-concurrency` parameter controls the maximum level of parallelism.

3. **Operational Steps**:
    For each repository, the following actions are performed:
    - Clone (`git clone`)
    - Create a random branch and switch to it (`git checkout -b <branch>`)
    - Create and add a random file (`git add`)
    - Commit changes (`git commit -m "Test commit"`)
    - Push the branch to the remote repository (`git push -u origin <branch>`)

4. **Logging and Statistics**:
    - Detailed logs are recorded for each repository's operations, including timing for each Git command.
    - After all workflows complete, the tool summarizes results and prints them to the console.

5. **Temporary Directory Cleanup**:
    - Once testing is complete, the `git-bench-tmp/` subdirectory is automatically cleaned up to maintain a tidy workspace.

---

## Notes

1. **Remote Repository Access**:
    - Ensure the provided repository URL is valid and that you have push permissions (preferably to a test repository).

2. **Push Branch Considerations**:
    - Each execution pushes multiple random branches to the remote repository. Ensure the repository allows write operations.
    - Clean up unused branches in the remote repository after testing.

3. **Resource Usage**:
    - Higher values for `-concurrency` may consume significant CPU and network bandwidth. Adjust this based on your hardware capabilities.

4. **Debug Logs**:
    - All logs are printed to standard output, making debugging simpler.

---

## Example Scenarios

### Testing Git Performance

Use high concurrency and repeated executions to assess the response time of a remote repository under different network conditions. For example:

```bash
./git-bench -repo=https://github.com/example/repo.git -concurrency=10 -count=20
```

By observing the time taken for `clone` and `push`, you can identify potential network latency or repository performance issues.

### Automating Repository Workflow Validation

In test environments, use `git-bench` to simulate workflows involving creating and pushing new branches and files, ensuring the environment is stable under load.

---

## License

[MIT](./LICENSE)
