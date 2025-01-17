package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// generateUUID 生成一个随机 UUID
func generateUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		log.Fatalf("Failed to generate UUID: %v", err)
	}
	return hex.EncodeToString(uuid)
}

// runGitCommand 在指定目录中运行 git 命令，并测量其执行时间
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

// processRepo 执行单个工作流操作，即克隆、checkout、add、commit 和 push
func processRepo(repoURL, baseDir string, index int) string {
	dir := filepath.Join(baseDir, fmt.Sprintf("repo_%d", index))

	var logBuffer strings.Builder

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		logBuffer.WriteString(fmt.Sprintf("Failed to create directory %s: %v\n", dir, err))
		return logBuffer.String()
	}

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

	return logBuffer.String()
}

func main() {
	repoURL := flag.String("repo", "", "Git repository URL")
	concurrency := flag.Int("concurrency", 1, "Number of concurrent processes")
	count := flag.Int("count", 1, "Number of iterations")
	flag.Parse()

	if *repoURL == "" {
		log.Fatalf("Repository URL is required")
	}

	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// 构造在当前目录下的临时目录
	baseDir := filepath.Join(currentDir, "fs-git-benchmark-tmp")
	err = os.MkdirAll(baseDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create base directory %s: %v", baseDir, err)
	}
	defer os.RemoveAll(baseDir) // 程序结束时删除临时目录

	var wg sync.WaitGroup
	workChan := make(chan int, *concurrency)

	startTime := time.Now()
	log.Printf("Starting git speed test with concurrency=%d, count=%d", *concurrency, *count)

	results := make([]string, *count)
	var resultsMu sync.Mutex

	for i := 0; i < *count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			workChan <- index
			defer func() { <-workChan }()

			logs := processRepo(*repoURL, baseDir, index)

			resultsMu.Lock()
			results[index] = logs
			resultsMu.Unlock()
		}(i)
	}

	wg.Wait()
	elapsedTime := time.Since(startTime)

	// 打印所有仓库的日志
	for i, logs := range results {
		log.Printf("Logs for repo_%d:\n%s", i, logs)
	}

	log.Printf("Completed test in %s", elapsedTime)
}
