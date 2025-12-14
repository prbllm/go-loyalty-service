package test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func findFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func waitForService(ctx context.Context, url string, timeout time.Duration) error {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for service at %s", url)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url+"/api/user/register", nil)
			if err != nil {
				continue
			}

			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode < 500 {
					time.Sleep(500 * time.Millisecond)
					return nil
				}
			}
		}
	}
}

func buildBinary() (string, error) {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..", "..")

	goModPath := filepath.Join(projectRoot, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		return "", fmt.Errorf("failed to find project root (go.mod not found at %s): %w", projectRoot, err)
	}

	binaryPath := filepath.Join(projectRoot, "testbin", "gophermart")

	if err := os.MkdirAll(filepath.Dir(binaryPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create testbin directory: %w", err)
	}

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/gophermart")
	cmd.Dir = projectRoot
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "CGO_ENABLED=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to build binary: %w, output: %s", err, string(output))
	}

	return binaryPath, nil
}

type process struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

func startProcess(ctx context.Context, binaryPath string, env []string, args ...string) (*process, error) {
	procCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(procCtx, binaryPath, args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	return &process{
		cmd:    cmd,
		cancel: cancel,
	}, nil
}

func (p *process) Stop() error {
	if p == nil {
		return nil
	}
	p.cancel()
	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Kill()
		return p.cmd.Wait()
	}
	return nil
}

func makeHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: false,
		},
	}
}

func startGophermart(ctx context.Context, binaryPath, dbURI, accrualURL string) (*process, int, error) {
	gophermartPort, err := findFreePort()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find port for gophermart: %w", err)
	}

	env := []string{
		fmt.Sprintf("RUN_ADDRESS=:%d", gophermartPort),
		fmt.Sprintf("DATABASE_URI=%s", dbURI),
		fmt.Sprintf("ACCRUAL_SYSTEM_ADDRESS=%s", accrualURL),
	}

	proc, err := startProcess(ctx, binaryPath, env)
	if err != nil {
		return nil, 0, err
	}

	baseURL := fmt.Sprintf("http://localhost:%d", gophermartPort)
	if err := waitForService(ctx, baseURL, 15*time.Second); err != nil {
		proc.Stop()
		return nil, 0, fmt.Errorf("gophermart failed to start: %w", err)
	}

	time.Sleep(1 * time.Second)

	return proc, gophermartPort, nil
}

func startAccrualMock(ctx context.Context) (*AccrualMock, *process, error) {
	mock, err := NewAccrualMock()
	if err != nil {
		return nil, nil, err
	}

	stopCtx, cancel := context.WithCancel(ctx)
	go func() {
		if err := mock.Start(); err != nil && err != http.ErrServerClosed {
			cancel()
		}
	}()

	if err := waitForService(ctx, mock.URL(), 5*time.Second); err != nil {
		mock.Stop(context.Background())
		return nil, nil, fmt.Errorf("accrual mock failed to start: %w", err)
	}

	proc := &process{
		cancel: func() {
			cancel()
			mock.Stop(stopCtx)
		},
	}

	return mock, proc, nil
}
