package server

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/peer"
)

var addr = [6]byte{127, 0, 0, 1, 0x1F, 0x90}

func TestListenAddr(t *testing.T) {

	s := NewServer(&addr)

	// Create a channel to receive errors
	errCh := make(chan error, 1)
	t.Log(convertAddrToString(&addr))

	go func() {
		defer close(errCh)
		conn, err := net.Dial(protocol, convertAddrToString(&addr))
		if err != nil {
			errCh <- fmt.Errorf("Failed to connect to server: %v", err)
			return
		}
		conn.Close()
	}()

	s.ln.Accept()
	s.ln.Close()

	if err := <-errCh; err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(addr[:]), 6)
	assert.Equal(t, convertAddrToString(&addr), "127.0.0.1:8080")
}

func TestHandleServer(t *testing.T) {
	server := NewServer(&addr)

	mockConn := &net.TCPConn{} // Replace with a mock TCPConn for testing
	server.TCPConn = mockConn

	go HandleServer(server)

	// Simulate reading/writing data to the connection, depending on how you mock net.TCPConn
}

func TestBroadcastVersionToPeers(t *testing.T) {
	server := NewServer(&addr)

	// Add a peer to the server
	peerAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	p := &peer.Peer{Addr: peerAddr}
	server.addPeer(p, server.getVersion())

	// Mock sendVersionToPeer
	sendVersionCalled := false

	server.broadcastVersionToPeers()

	if !sendVersionCalled {
		t.Error("Expected sendVersionToPeer to be called during broadcast")
	}
}

func TestGetVersion(t *testing.T) {
	server := NewServer(&addr)
	version := server.getVersion()

	if version.Version != 1 {
		t.Errorf("Expected version 1, got %d", version.Version)
	}

	if len(version.PeerList) != 0 {
		t.Errorf("Expected no peers in version peer list, got %d", len(version.PeerList))
	}
}

func TestConvertAddrToByteArray(t *testing.T) {
	server := NewServer(&addr)
	addr := "127.0.0.1:8080"
	expected := &[6]byte{127, 0, 0, 1, 0x1F, 0x90} // 127.0.0.1:8080

	result := server.convertAddrToByteArray(addr)

	if *result != *expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCanConnect(t *testing.T) {
	server := NewServer(&addr)
	addr := &[6]byte{127, 0, 0, 1, 0x1F, 0x91} // 127.0.0.1:8081

	canConnect := server.canConnect(addr)
	if !canConnect {
		t.Error("Expected server to be able to connect to a new peer")
	}

	// Simulate an existing peer
	peerAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8081")
	p := &peer.Peer{Addr: peerAddr}
	server.addPeer(p, server.getVersion())

	canConnect = server.canConnect(addr)
	if canConnect {
		t.Error("Expected server to not connect to an already added peer")
	}
}

func TestRemovePeer(t *testing.T) {
	server := NewServer(&addr)
	peerAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	p := &peer.Peer{Addr: peerAddr}
	server.addPeer(p, server.getVersion())

	server.removePeer(p)

	if len(server.peers) != 0 {
		t.Errorf("Expected 0 peers, got %d", len(server.peers))
	}
}
func TestAddPeer(t *testing.T) {
	server := NewServer(&addr)
	peerAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	p := &peer.Peer{Addr: peerAddr}

	server.addPeer(p, server.getVersion())

	if len(server.peers) != 1 {
		t.Errorf("Expected 1 peer, got %d", len(server.peers))
	}

	if _, exists := server.peers[p]; !exists {
		t.Error("Expected peer to be in the server's peer map")
	}
}
