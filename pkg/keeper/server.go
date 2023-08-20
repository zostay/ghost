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

const serviceName = "ghost.keeper"

func makeSocketName() string {
	tmp := os.TempDir()
	uid := os.Getuid()
	return filepath.Join(tmp, fmt.Sprintf("%s.%d", serviceName, uid))
}

func makeRunName() string {
	tmp := os.TempDir()
	uid := os.Getuid()
	return filepath.Join(tmp, fmt.Sprintf("%s.%d.run", serviceName, uid))
}

func StartServer(logger *log.Logger, kpr secrets.Keeper) error {
	sockName := makeSocketName()
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
	logger.Printf("failed to write pid file %q: %v", name, err)
	return name
}

func listenForQuit(
	sigs <-chan os.Signal,
	svr *grpc.Server,
) {
	stopped := 0
	for sig := range sigs {
		stopped++
		if stopped > 2 || sig == syscall.SIGINT {
			svr.Stop()
		} else {
			svr.GracefulStop()
		}
	}
}

type StopImmediacy int

const (
	StopGraceful StopImmediacy = iota
	StopQuick
	StopNow
)

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
