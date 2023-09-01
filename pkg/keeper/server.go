package keeper

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"google.golang.org/grpc"

	"github.com/zostay/ghost/pkg/secrets"
	"github.com/zostay/ghost/pkg/secrets/http"
)

func makeRunName() string {
	tmp := os.TempDir()
	uid := os.Getuid()
	return filepath.Join(tmp, fmt.Sprintf("%s.%d.run", http.ServiceName, uid))
}

// StartServer starts the keeper server. As of this writing, it will always be
// configured to run in an automatically named unix socket in the system's temp
// directory. It will also write a pid file to the same directory.
func StartServer(logger *log.Logger, kpr secrets.Keeper) error {
	sockName := http.MakeHttpServerSocketName()
	sock, err := net.Listen("unix", sockName)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket %q: %w", sockName, err)
	}
	defer func() {
		_ = sock.Close()
		_ = os.Remove(sockName)
	}()

	gracefulQuitter := make(chan os.Signal, 3)
	signal.Notify(gracefulQuitter, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)

	pidFile := makePidFile(logger)
	defer func() { _ = os.Remove(pidFile) }()

	svr := http.NewServer(kpr)
	grpcServer := grpc.NewServer()
	http.RegisterKeeperServer(grpcServer, svr)
	go listenForQuit(gracefulQuitter, grpcServer)
	err = grpcServer.Serve(sock)
	if err != nil {
		return fmt.Errorf("grpc server quit with error: %w", err)
	}

	return nil
}

func makePidFile(logger *log.Logger) string {
	name := makeRunName()
	pid := fmt.Sprintf("%d", os.Getpid())
	err := os.WriteFile(name, []byte(pid), 0o600)
	if err != nil {
		logger.Printf("failed to write pid file %q: %v", name, err)
	}
	return name
}

func listenForQuit(
	sigs <-chan os.Signal,
	svr *grpc.Server,
) {
	stopped := 0
	for sig := range sigs {
		stopped++
		if stopped > 2 || sig == syscall.SIGINT || sig == syscall.SIGQUIT {
			svr.Stop()
		} else {
			svr.GracefulStop()
		}
	}
}

// StopImmediacy is used to indicate how quickly the server should be stopped.
type StopImmediacy int

const (
	StopGraceful StopImmediacy = iota // stop eventually (SIGHUP)
	StopQuick                         // stop soon (SIGQUIT)
	StopNow                           // stop now (SIGKILL)
)

// StopServer stops the keeper server. The given immediacy indicates how quickly
// the server should be stopped.
func StopServer(immediacy StopImmediacy) error {
	pidFile := makeRunName()
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("unable to locate pid file %q; you'll have to kill the process manually: %w", pidFile, err)
	}

	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		return fmt.Errorf("unable to read pid file %q; you'll have to kill the process manually: %w", pidFile, err)
	}

	var sig syscall.Signal
	switch immediacy {
	case StopGraceful:
		sig = syscall.SIGHUP
	case StopNow:
		sig = syscall.SIGKILL
	case StopQuick:
		sig = syscall.SIGQUIT
	default:
		sig = syscall.SIGHUP
	}
	err = syscall.Kill(pid, sig)
	if err != nil {
		return fmt.Errorf("failed to send pid %d a signal: %w", pid, err)
	}

	return nil
}

// CheckServer checks if the server is alive. Returns an error if it is night.
func CheckServer() error {
	pidFile := makeRunName()
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("unable to locate pid file %q; you'll have to kill the process manually: %w", pidFile, err)
	}

	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		return fmt.Errorf("unable to read pid file %q; you'll have to kill the process manually: %w", pidFile, err)
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("unable to find process for pid %d: %w", pid, err)
	}

	err = p.Signal(syscall.Signal(0))
	if err != nil {
		return fmt.Errorf("unable to verify process for pid %d: %w", pid, err)
	}

	return nil
}
