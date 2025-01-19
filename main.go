package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// generateUUID 生成一个 UUID
func generateUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		log.Fatalf("Failed to generate UUID: %v", err)
	}
	return hex.EncodeToString(uuid)
}

// runGitCommand 使用本地 git 命令执行 Git 操作
func runGitCommand(dir string, args ...string) (time.Duration, string, error) {
	start := time.Now()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)
	var logMsg string
	if err != nil {
		logMsg = fmt.Sprintf("Error running git %v: %v (Time: %s) | %s", args, err, duration, strings.TrimSpace(stderr.String()))
	} else {
		logMsg = fmt.Sprintf("Git command %v executed successfully in %s", args, duration)
	}

	return duration, logMsg, err
}

// createSSHAuth 使用私钥加载 SSH 认证
func createSSHAuth(privateKeyPath string) (transport.AuthMethod, error) {
	// 使用指定的私钥
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %v", err)
	}

	sshAuth, err := ssh.NewPublicKeys("git", key, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH auth from private key: %v", err)
	}

	return sshAuth, nil
}

// cloneUsingGoGit 使用 go-git 克隆仓库
func cloneUsingGoGit(repoURL, dir, privateKeyPath string) error {
	sshAuth, err := createSSHAuth(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create SSH auth: %v", err)
	}

	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL:      repoURL,
		Auth:     sshAuth,
		Progress: os.Stdout,
	})

	return err
}

// checkoutNewBranch 使用 go-git 切换分支
func checkoutNewBranch(repoPath, branchName string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	workTree, err := repo.Worktree()
	if err != nil {
		return err
	}

	// 创建并切换到新分支
	ref := plumbing.NewBranchReferenceName(branchName)
	err = workTree.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: ref,
	})

	return err
}

// addFileUsingGoGit 使用 go-git 添加文件
func addFileUsingGoGit(repoPath string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	workTree, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = workTree.Add(".")
	return err
}

// commitUsingGoGit 使用 go-git 提交更改
func commitUsingGoGit(repoPath, message string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	workTree, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = workTree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "git-bench",
			Email: "git-bench@example.com",
			When:  time.Now(),
		},
	})

	return err
}

// pushUsingGoGit 使用 go-git 推送分支
func pushUsingGoGit(repoPath, privateKeyPath string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	sshAuth, err := createSSHAuth(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create SSH auth: %v", err)
	}

	err = repo.Push(&git.PushOptions{
		Auth:     sshAuth,
		Progress: os.Stdout,
	})
	return err
}

// processRepo 根据用户选择使用本地 git 或 go-git 完成流程
func processRepo(repoURL, baseDir string, index int, useGoGit bool, privateKeyPath string) string {
	dir := filepath.Join(baseDir, fmt.Sprintf("repo_%d", index))
	var logBuffer strings.Builder
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		logBuffer.WriteString(fmt.Sprintf("Failed to create directory %s: %v\n", dir, err))
		return logBuffer.String()
	}

	if useGoGit {
		// 克隆仓库
		logBuffer.WriteString(fmt.Sprintf("Cloning repository %s into %s using go-git...\n", repoURL, dir))
		start := time.Now()
		err := cloneUsingGoGit(repoURL, dir, privateKeyPath)
		elapsed := time.Since(start)
		logBuffer.WriteString(fmt.Sprintf("Cloning completed in %s\n", elapsed))
		if err != nil {
			logBuffer.WriteString(fmt.Sprintf("Failed to clone repo: %v\n", err))
			return logBuffer.String()
		}

		// 创建新分支
		branchName := generateUUID()
		start = time.Now()
		logBuffer.WriteString(fmt.Sprintf("Checking out to new branch %s...\n", branchName))
		err = checkoutNewBranch(dir, branchName)
		elapsed = time.Since(start)
		logBuffer.WriteString(fmt.Sprintf("Checkout completed in %s\n", elapsed))
		if err != nil {
			logBuffer.WriteString(fmt.Sprintf("Failed to checkout branch: %v\n", err))
			return logBuffer.String()
		}

		// 创建新文件
		fileName := filepath.Join(dir, generateUUID()+".txt")
		start = time.Now()
		logBuffer.WriteString(fmt.Sprintf("Creating new file %s...\n", fileName))
		err = os.WriteFile(fileName, []byte("This is a test file"), 0644)
		elapsed = time.Since(start)
		logBuffer.WriteString(fmt.Sprintf("File creation completed in %s\n", elapsed))
		if err != nil {
			logBuffer.WriteString(fmt.Sprintf("Failed to write file: %v\n", err))
			return logBuffer.String()
		}

		// 添加文件到索引
		start = time.Now()
		logBuffer.WriteString(fmt.Sprintf("Staging file %s using go-git...\n", fileName))
		err = addFileUsingGoGit(dir)
		elapsed = time.Since(start)
		logBuffer.WriteString(fmt.Sprintf("Staging completed in %s\n", elapsed))
		if err != nil {
			logBuffer.WriteString(fmt.Sprintf("Failed to add file: %v\n", err))
			return logBuffer.String()
		}

		// 提交更改
		start = time.Now()
		logBuffer.WriteString(fmt.Sprintf("Committing changes...\n"))
		err = commitUsingGoGit(dir, "Test commit")
		elapsed = time.Since(start)
		logBuffer.WriteString(fmt.Sprintf("Commit completed in %s\n", elapsed))
		if err != nil {
			logBuffer.WriteString(fmt.Sprintf("Failed to commit: %v\n", err))
			return logBuffer.String()
		}

		// 推送到远程
		start = time.Now()
		logBuffer.WriteString(fmt.Sprintf("Pushing branch to remote using go-git...\n"))
		err = pushUsingGoGit(dir, privateKeyPath)
		elapsed = time.Since(start)
		logBuffer.WriteString(fmt.Sprintf("Pushing completed in %s\n", elapsed))
		if err != nil {
			logBuffer.WriteString(fmt.Sprintf("Failed to push: %v\n", err))
			return logBuffer.String()
		}
	} else {
		// Clone the repository
		logBuffer.WriteString(fmt.Sprintf("Cloning repository %s into %s\n", repoURL, dir))
		_, logMsg, err := runGitCommand("", "clone", repoURL, dir)
		logBuffer.WriteString(logMsg + "\n")
		if err != nil {
			return logBuffer.String()
		}

		// Create and checkout a new branch
		branchName := generateUUID()
		logBuffer.WriteString(fmt.Sprintf("Checking out to new branch %s in %s\n", branchName, dir))
		_, logMsg, err = runGitCommand(dir, "checkout", "-b", branchName)
		logBuffer.WriteString(logMsg + "\n")
		if err != nil {
			return logBuffer.String()
		}

		// Create a new random file
		fileName := filepath.Join(dir, generateUUID()+".txt")
		logBuffer.WriteString(fmt.Sprintf("Creating new file %s\n", fileName))
		err = os.WriteFile(fileName, []byte("This is a test file"), 0644)
		if err != nil {
			logBuffer.WriteString(fmt.Sprintf("Failed to write file: %v\n", err))
			return logBuffer.String()
		}

		// Stage the new file
		logBuffer.WriteString(fmt.Sprintf("Staging file %s\n", fileName))
		_, logMsg, err = runGitCommand(dir, "add", ".")
		logBuffer.WriteString(logMsg + "\n")
		if err != nil {
			return logBuffer.String()
		}

		// Commit the change
		logBuffer.WriteString(fmt.Sprintf("Committing changes in %s\n", dir))
		_, logMsg, err = runGitCommand(dir, "commit", "-m", "Test commit")
		logBuffer.WriteString(logMsg + "\n")
		if err != nil {
			return logBuffer.String()
		}

		// Push the new branch
		logBuffer.WriteString(fmt.Sprintf("Pushing branch %s from %s\n", branchName, dir))
		_, logMsg, err = runGitCommand(dir, "push", "-u", "origin", branchName)
		logBuffer.WriteString(logMsg + "\n")
		if err != nil {
			return logBuffer.String()
		}
	}

	return logBuffer.String()
}

func main() {
	repoURL := flag.String("repo", "", "Git repository URL")
	concurrency := flag.Int("concurrency", 1, "Number of concurrent processes")
	count := flag.Int("count", 1, "Number of iterations")
	useGoGit := flag.Bool("use-go-git", false, "Use go-git library instead of local git commands")
	sshKey := flag.String("ssh-key", "", "Path to SSH private key (leave empty to use SSH agent)")
	flag.Parse()

	if *repoURL == "" {
		log.Fatalf("Repository URL is required")
	}

	if *useGoGit && *sshKey == "" {
		log.Fatalf("SSH key is required")
	}

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	baseDir := filepath.Join(currentDir, "git-bench-tmp")
	err = os.MkdirAll(baseDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create base directory %s: %v", baseDir, err)
	}
	defer os.RemoveAll(baseDir)

	var wg sync.WaitGroup
	workChan := make(chan int, *concurrency)
	startTime := time.Now()
	log.Printf("Starting git bench with concurrency=%d, count=%d, useGoGit=%v sshKey=%s", *concurrency, *count, *useGoGit, *sshKey)

	results := make([]string, *count)
	var resultsMu sync.Mutex

	for i := 0; i < *count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			workChan <- index
			defer func() { <-workChan }()

			logs := processRepo(*repoURL, baseDir, index, *useGoGit, *sshKey)
			resultsMu.Lock()
			results[index] = logs
			resultsMu.Unlock()
		}(i)
	}

	wg.Wait()
	elapsedTime := time.Since(startTime)

	for i, logs := range results {
		log.Printf("Logs for repo_%d:\n%s", i, logs)
	}

	log.Printf("Completed git bench in %s", elapsedTime)
}
