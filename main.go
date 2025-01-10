// Copyright 2025 Aristore. All rights reserved.
//
// MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

// 定义配置结构体
type Replacement struct {
	Pattern     string `toml:"pattern"`
	Replacement string `toml:"replacement"`
}

type Config struct {
	PythonPath           string        `toml:"python_path"`
	EnableRealtimeOutput bool          `toml:"enable_realtime_output"`
	Replacements         []Replacement `toml:"replacements"`
}

// 从与可执行文件相同的目录加载配置文件
func loadConfig(executablePath string) (Config, error) {
	configPath := filepath.Join(executablePath, "config.toml")
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return config, fmt.Errorf("error loading config file: %v", err)
	}
	return config, nil
}

// 创建自定义输出处理函数
func createCustomOutput(replacements []Replacement) func(string) string {
	var compiledReplacements []*regexp.Regexp
	var replacementStrings []string
	for _, r := range replacements {
		compiledPattern, err := regexp.Compile(r.Pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid regex pattern in config: %v\n", err)
			continue
		}
		compiledReplacements = append(compiledReplacements, compiledPattern)
		replacementStrings = append(replacementStrings, r.Replacement)
	}
	return func(text string) string {
		for i, pattern := range compiledReplacements {
			text = pattern.ReplaceAllString(text, replacementStrings[i])
		}
		return text
	}
}

// 处理子进程的输出流
func processStream(stream *bufio.Scanner, customWrite func(string), wg *sync.WaitGroup) {
	defer wg.Done()
	for stream.Scan() {
		customWrite(stream.Text() + "\n")
	}
	if err := stream.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stream: %v\n", err)
	}
}

// 处理用户输入并发送给子进程的标准输入
func handleStdin(stdinPipe io.WriteCloser) {
	defer stdinPipe.Close() // 确保在函数结束时关闭标准输入

	// 创建一个扫描器来读取用户的输入
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// 将用户的输入发送到子进程的标准输入
		fmt.Fprintln(stdinPipe, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
	}
}

func main() {
	// 获取当前可执行文件的绝对路径
	ex, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get executable path: %v\n", err)
		return
	}
	exPath := filepath.Dir(ex)

	// 加载配置文件
	config, err := loadConfig(exPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("Usage: python script.py [args...]")
		return
	}

	// Create a function for custom output processing
	customWrite := createCustomOutput(config.Replacements)

	cmd := exec.Command(config.PythonPath, os.Args[1:]...)

	// 设置环境变量，确保UTF-8编码
	cmd.Env = append(os.Environ(), "PYTHONIOENCODING=utf8", "LANG=zh_CN.UTF-8")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get stdout pipe: %v\n", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get stderr pipe: %v\n", err)
		return
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get stdin pipe: %v\n", err)
		return
	}

	scannerOut := bufio.NewScanner(stdout)
	scannerErr := bufio.NewScanner(stderr)

	var wg sync.WaitGroup
	wg.Add(2)

	if config.EnableRealtimeOutput {
		err = cmd.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start command: %v\n", err)
			return
		}

		go processStream(scannerOut, func(text string) { fmt.Fprint(os.Stdout, customWrite(text)) }, &wg)
		go processStream(scannerErr, func(text string) { fmt.Fprint(os.Stderr, customWrite(text)) }, &wg)

		// 启动一个goroutine来处理用户输入并发送给子进程
		go handleStdin(stdin)

		wg.Wait()

		// 等待命令完成
		err = cmd.Wait()
		if err != nil {
			// fmt.Fprintf(os.Stderr, "Command finished with error: %v\n", err)
			return
		}
	} else {
		// Start the command and wait for it to complete.
		if err = cmd.Run(); err != nil {
			// fmt.Fprintf(os.Stderr, "Command finished with error: %v\n", err)
			return
		}

		// In non-realtime mode, we can simply capture the combined output and apply the replacements at once.
		output, err := cmd.CombinedOutput()
		if err != nil {
			// fmt.Fprintf(os.Stderr, "Command finished with error: %v\n", err)
			return
		}
		fmt.Print(customWrite(strings.TrimSpace(string(output))))
	}
}
