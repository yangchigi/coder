package ctrlsock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/xerrors"

	"cdr.dev/slog"
)

type Server struct {
	logger   slog.Logger
	handlers Handlers
	ln       net.Listener
	wg       sync.WaitGroup
	done     chan struct{}
	authKey  string
}

func (s *Server) Addr() net.Addr {
	return s.ln.Addr()
}

func (s *Server) AuthKey() string {
	return s.authKey
}

type Handlers struct {
	SetEnv func(key, value string)
}

func New(logger slog.Logger, runDir string, handlers Handlers) (*Server, error) {
	addr := filepath.Join(runDir, "agent.sock")

	err := os.Remove(addr)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, xerrors.Errorf("remove existing socket failed: %w", err)
	}

	ln, err := net.Listen("unix", addr)
	if err != nil {
		return nil, err
	}

	authKey, err := generateAuthKey()
	if err != nil {
		return nil, err
	}

	s := &Server{
		logger:   logger.Named("ctrlsock"),
		handlers: handlers,
		ln:       ln,
		done:     make(chan struct{}),
		authKey:  authKey,
	}
	go s.acceptLoop()

	return s, nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			select {
			case <-s.done:
				// The listener was closed, so we're done.
				return
			default:
				s.logger.Error(context.Background(), "accept connection failed", "err", err)
			}
		} else {
			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	// Check the authentication key.
	if err := s.handleAuth(conn); err != nil {
		s.logger.Error(context.Background(), "authentication failed", "err", err)
		return
	}

	// Handle commands.
	for {
		cmdByte, err := readByte(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			s.logger.Error(context.Background(), "read command type failed", "err", err)
			return
		}
		cmdType := Command(cmdByte)
		logger := s.logger.With(slog.F("command", cmdType.String()))

		switch cmdType {
		case SetEnv:
			key, value, err := readSetEnv(conn)
			if err != nil {
				logger.Error(context.Background(), "handle command failed", "err", err)
			}
			logger.Info(context.Background(), "calling command with input", "key", key, "value_length", len(value))
			s.handlers.SetEnv(key, value)
		default:
			s.logger.Error(context.Background(), "unknown command, closing connection")
			return
		}
	}
}

func generateAuthKey() (string, error) {
	key := make([]byte, 16)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

func (s *Server) handleAuth(conn net.Conn) error {
	key, err := readString(conn)
	if err != nil {
		return err
	}
	if key != s.authKey {
		return xerrors.Errorf("invalid auth key: %s", key)
	}
	return nil
}

func (s *Server) Close() error {
	close(s.done)
	err := s.ln.Close()
	s.wg.Wait()
	return err
}
