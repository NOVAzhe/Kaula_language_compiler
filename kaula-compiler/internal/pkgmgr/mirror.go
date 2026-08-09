package pkgmgr

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// isValidMirrorURL 验证镜像 URL 是否安全（仅允许 HTTP/HTTPS，禁止内网地址）
func isValidMirrorURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// 仅允许 http 和 https 协议
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}

	// 禁止包含用户信息（认证凭据）
	if u.User != nil {
		return fmt.Errorf("URL must not contain user credentials")
	}

	// 解析主机名
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("empty host")
	}

	// 禁止 IP 文字和 localhost
	if strings.HasPrefix(host, "127.") || host == "localhost" || host == "::1" || host == "0.0.0.0" {
		return fmt.Errorf("forbidden host: %s", host)
	}

	// 尝试解析 IP，检查是否为内网地址
	ips, err := net.LookupHost(host)
	if err == nil {
		for _, ip := range ips {
			parsed := net.ParseIP(ip)
			if parsed != nil {
				if parsed.IsPrivate() || parsed.IsLoopback() || parsed.IsLinkLocalUnicast() || parsed.IsUnspecified() {
					return fmt.Errorf("forbidden private IP: %s", ip)
				}
			}
		}
	}

	return nil
}

// MirrorTestResult 镜像测速结果
type MirrorTestResult struct {
	URL     string
	Latency time.Duration // 延迟（RTT）
	Speed   float64       // 速度（bytes/sec）
	Success bool
	Error   error
}

// TestMirror 测试单个镜像的延迟和速度
// 发送 Range 请求下载前 64KB 来测量
func TestMirror(url string, timeout time.Duration) MirrorTestResult {
	result := MirrorTestResult{URL: url}

	// 安全校验：防止 SSRF
	if err := isValidMirrorURL(url); err != nil {
		result.Error = fmt.Errorf("SSRF check failed: %w", err)
		return result
	}

	client := &http.Client{Timeout: timeout}

	start := time.Now()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		result.Error = err
		return result
	}
	// 只下载前 64KB 来测速
	req.Header.Set("Range", "bytes=0-65535")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		result.Error = fmt.Errorf("HTTP %d", resp.StatusCode)
		return result
	}

	// 读取 64KB 数据
	n, err := io.Copy(io.Discard, resp.Body)
	elapsed := time.Since(start)

	if err != nil && err != io.EOF {
		result.Error = err
		return result
	}

	result.Latency = elapsed
	result.Speed = float64(n) / elapsed.Seconds()
	result.Success = true
	return result
}

// SelectBestMirror 并发测试所有镜像，选择最快的
// 返回最佳镜像 URL 和详细测速结果
func SelectBestMirror(mirrors []string, timeout time.Duration) (string, []MirrorTestResult) {
	if len(mirrors) == 0 {
		return "", nil
	}
	if len(mirrors) == 1 {
		return mirrors[0], []MirrorTestResult{{URL: mirrors[0], Success: true}}
	}

	results := make([]MirrorTestResult, len(mirrors))
	var wg sync.WaitGroup

	fmt.Printf("  Testing %d mirrors...\n", len(mirrors))

	for i, mirror := range mirrors {
		wg.Add(1)
		go func(idx int, m string) {
			defer wg.Done()
			results[idx] = TestMirror(m, timeout)
			if results[idx].Success {
				fmt.Printf("  [OK] %s - %dms, %.1f KB/s\n",
					m, results[idx].Latency.Milliseconds(),
					results[idx].Speed/1024)
			} else {
				fmt.Printf("  [FAIL] %s - %v\n", m, results[idx].Error)
			}
		}(i, mirror)
	}

	wg.Wait()

	// 选择速度最快的镜像
	bestIdx := -1
	var bestSpeed float64
	for i, r := range results {
		if r.Success {
			if bestIdx == -1 || r.Speed > bestSpeed {
				bestIdx = i
				bestSpeed = r.Speed
			}
		}
	}

	if bestIdx == -1 {
		// 所有镜像都失败，返回第一个
		return mirrors[0], results
	}

	fmt.Printf("  Best mirror: %s (%.1f KB/s)\n", mirrors[bestIdx], bestSpeed/1024)
	return mirrors[bestIdx], results
}
