package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/NikoMalik/blockchain/proto"
	"github.com/NikoMalik/blockchain/server/p2p"
	"github.com/NikoMalik/mutex"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, 1024<<10)
	},
}

var _ proto.ServerServer = (*Server)(nil)

const protocol = "tcp"
const (
	defaultDialTimeout = 20 * time.Second

	discmixTimeout = 5 * time.Second

	defaultMaxPendingPeers = 50
	defaultDialRatio       = 3

	inboundThrottleTime = 30 * time.Second

	frameReadTimeout = 30 * time.Second

	frameWriteTimeout = 20 * time.Second
)

type Server struct {
	proto.UnimplementedServerServer
	*net.TCPConn
	mu    mutex.MutexExp
	ln    net.Listener
	peers map[*peer.Peer]*proto.Version

	logger     *zap.Logger
	grpc       *grpc.Server
	listenAddr [6]byte
	version    int32
}

func NewServer(listenAddr *[6]byte, opts ...grpc.ServerOption) *Server {
	addrStr := convertAddrToString(listenAddr)
	ln, err := net.Listen(protocol, addrStr)
	if err != nil {
		panic(err)
	}

	srv := &Server{
		listenAddr: *listenAddr,
		ln:         ln,
		version:    0x01, //1
		grpc:       grpc.NewServer(opts...),
		peers:      make(map[*peer.Peer]*proto.Version, defaultMaxPendingPeers),
		logger:     createLogger(),
	}

	proto.RegisterServerServer(srv.grpc, srv)

	return srv
}

func (s *Server) LoggerReturn() *zap.Logger {
	return s.logger
}

func createLogger() *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = ""
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder // Colorize the log level
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.DisableCaller = false
	config.DisableStacktrace = false
	config.Sampling = nil
	config.Encoding = "console"
	config.EncoderConfig = encoderCfg
	config.OutputPaths = []string{"stderr"}
	config.ErrorOutputPaths = []string{"stderr"}
	// config.InitialFields = map[string]interface{}{
	// 	"pid": os.Getpid(),
	// }

	return zap.Must(config.Build())
}
func (s *Server) addPeer(p *peer.Peer, v *proto.Version) {

	s.mu.Lock()
	s.peers[p] = v
	s.mu.Unlock()

	s.logger.Sugar().Infof("Peer added: %s", p.Addr.String())
	if len(s.peers) > 0 {
		go s.broadcastVersionToPeers()
	}
}

func (s *Server) broadcastVersionToPeers() {
	peers := s.getPeerList()

	pChan := make(chan *peer.Peer, len(peers))
	errChan := make(chan error, len(peers))

	var wg = &sync.WaitGroup{}

	for _, peerAddr := range peers {
		if !s.canConnect(s.convertAddrToByteArray(peerAddr)) {
			continue
		}
		wg.Add(1)
		go func(addr string) {
			if addr != peerAddr {
				s.logger.Error("Failed to dial peer", zap.String("addr", addr))
				syscall.Exit(1)
			}

			tcpAddr, err := net.ResolveTCPAddr(protocol, addr)
			if err != nil {
				s.logger.Sugar().Errorf("Failed to resolve address for peer %s: %v", addr, err)
				return
			}

			p := &peer.Peer{Addr: tcpAddr}

			pChan <- p
			wg.Done()
		}(peerAddr)
	}

	wg.Wait()

	go func() {
		for p := range pChan {

			err := s.sendVersionToPeer(p)
			if err != nil {
				errChan <- fmt.Errorf("failed to send version to peer %s: %v", p.Addr.String(), err)
			}
		}
		close(errChan)
	}()

	for err := range errChan {
		s.logger.Sugar().Error(err)
		// syscall.Exit(1)
	}
}

func (s *Server) canConnect(addr *[6]byte) bool {
	if s.getPeerList() == nil {
		return false
	}
	if s.listenAddr == *addr {
		return false
	}
	connected := s.getPeerList()

	for i := 0; i < len(connected); i++ {
		if addr == s.convertAddrToByteArray(connected[i]) {
			return false
		}
	}
	return true
}

func (s *Server) convertAddrToByteArray(addr string) *[6]byte {
	addrBytes := [6]byte{}       // no need to make a slice, just a plain array
	copy(addrBytes[:], addr[:6]) // only copy the first 6 bytes, the rest are ignored
	return &addrBytes
}
func (s *Server) getVersion() *proto.Version {
	return &proto.Version{
		Version:  s.version,
		Height:   0,
		PeerList: s.getPeerList(),
	}
}

func (s *Server) sendVersionToPeer(p *peer.Peer) error {

	peerAddr := p.Addr.String()

	conn, err := grpc.Dial(peerAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		s.logger.Sugar().Errorf("Failed to connect to peer %s: %v", peerAddr, err)
		return err
	}

	client := proto.NewServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)

	ch := make(chan struct{})
	go func() {
		<-ch
		conn.Close()
		cancel()
	}()

	response, err := client.HandShake(ctx, s.getVersion())
	if err != nil {
		s.logger.Sugar().Errorf("Failed to send version to peer %s: %v", peerAddr, err)
		return err
	}

	s.logger.Sugar().Infof("Version sent successfully to peer %s. Response version: %v", peerAddr, response.Version)

	ch <- struct{}{}

	return nil
}

func (s *Server) removePeer(c *peer.Peer) {
	s.mu.Lock()
	delete(s.peers, c)
	s.mu.Unlock()
}

func (s *Server) GetId(ctx context.Context, proto *proto.Block) (*proto.Result, error) {
	return nil, nil
}

func GetIpAddresses() []string {

	name, err := os.Hostname()
	if err != nil {

		return nil
	}

	addrs, err := net.LookupHost(name)
	if err != nil {

		return nil
	}

	return addrs
}

func networkError(err error, server *Server) {
	server.logger.Error("Network error: ", zap.Error(err))
	if err == io.EOF || err != nil {
		server.TCPConn.Close()
	}
}

func HandleServer(server *Server) {
	buf := bufferPool.Get().([]byte)

	for {

		n, err := server.TCPConn.Read(buf)
		if err != nil {
			networkError(err, server)
			if err == io.EOF {
				server.logger.Info("Connection closed by peer")
				break
			}
			continue
		}

		m := new(p2p.Msg)
		err = m.UnmarshalBinary(buf[:n])
		if err != nil {
			networkError(err, server)
			continue
		}

		m.Reply = make(chan p2p.Msg)

		go func(c chan p2p.Msg) {
			for replyMsg := range c {
				b, _ := replyMsg.MarshalBinary()
				_, err := server.TCPConn.Write(b)
				if err != nil {
					server.logger.Error("Error writing to connection", zap.Error(err))
					break
				}
			}
			close(c)
		}(m.Reply)

		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			server.logger.Warn("Connection timeout", zap.Error(opErr))
			continue
		}

		networkError(err, server)
		break

	}

	bufferPool.Put(&buf)
	server.TCPConn.Close()
}
func convertAddrToString(addr *[6]byte) string {
	return fmt.Sprintf("%d.%d.%d.%d:%d",
		addr[0], addr[1], addr[2], addr[3],
		int(addr[4])<<8|int(addr[5]))
}

func (s *Server) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Result, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return &proto.Result{
			Code:    1,
			Message: "Peer not found",
		}, nil
	}
	s.logger.Sugar().Info("Transaction received", peer)
	return &proto.Result{
		Code:    0,
		Message: "Success",
	}, nil

}

func (s *Server) getPeerList() []string {

	s.mu.Lock()
	peers := make([]string, 0, len(s.peers))
	for p := range s.peers {
		peers = append(peers, p.Addr.String())
	}
	s.mu.Unlock()

	return peers

}

// example addr = &[6]byte{127, 0, 0, 1, 0x1F, 0x90}// 127.0.0.1:8080
func (s *Server) Start() error {
	go func() {
		for {

			_, err := s.ln.Accept()
			if err != nil {
				s.logger.Error("Error accepting connection", zap.Error(err))
				continue
			}

			go HandleServer(s)
		}
	}()

	if err := s.grpc.Serve(s.ln); err != nil {
		if !errors.Is(err, grpc.ErrServerStopped) {
			panic(err)
		}
	}

	return nil
}

func (s *Server) Stop() error {
	// Create a channel to receive an OS interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for the signal
	<-quit

	// Close all active connections and clean up
	// Additional cleanup logic if needed

	// Stop gRPC server
	s.grpc.GracefulStop()

	// Close the listener
	s.ln.Close()

	// Close the TCP connection
	s.TCPConn.Close()

	// Remove all peers
	s.mu.Lock()
	for p := range s.peers {
		s.removePeer(p)
	}
	s.mu.Unlock()

	s.logger.Info("Server stopped")
	return nil
}

func (s *Server) GrpcReturn() *grpc.Server {
	return s.grpc
}

func (s *Server) HandShake(ctx context.Context, version *proto.Version) (*proto.Version, error) {

	peer, _ := peer.FromContext(ctx)

	func() { s.addPeer(peer, s.getVersion()) }()

	return s.getVersion(), nil

}
