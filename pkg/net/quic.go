package net

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"time"

	"github.com/edenzhong7/xrpc/pkg/algs"

	quic "github.com/lucas-clemente/quic-go"
)

const (
	proto = "xrpc-quic"
)

var (
	quicDialer Dialer = func(ctx context.Context, addr string) (conn Conn, err error) {
		tlsConf := &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{proto},
		}
		session, err := quic.DialAddr(addr, tlsConf, nil)
		if err != nil {
			return nil, err
		}
		stream, err := session.OpenStreamSync(context.Background())
		if err != nil {
			return nil, err
		}
		return newQuicConnection(stream, session.LocalAddr().String(), session.RemoteAddr().String())
	}
	_ Listener = &QUICListener{}
)

func init() {
	RegisterDialer(QUIC, quicDialer)
	RegisterListenerBuilder(QUIC, newQuicListener)
}

var quicConfig = &quic.Config{
	MaxIncomingStreams:                    1000,
	MaxIncomingUniStreams:                 -1,              // disable unidirectional streams
	MaxReceiveStreamFlowControlWindow:     3 * (1 << 20),   // 3 MB
	MaxReceiveConnectionFlowControlWindow: 4.5 * (1 << 20), // 4.5 MB
	AcceptToken: func(clientAddr net.Addr, cookie *quic.Token) bool {
		return true
	},
	KeepAlive: true,
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{proto},
	}
}

func newQuicListener(ctx context.Context, addr string) (Listener, error) {
	quicListener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	maxBufSize := 512
	if err != nil {
		return nil, err
	}
	return &QUICListener{
		ctx:        ctx,
		addr:       &XAddr{QUIC, addr},
		listener:   quicListener,
		maxBufSize: maxBufSize,
	}, nil
}

type QUICListener struct {
	ctx        context.Context
	listener   quic.Listener
	addr       Addr
	maxBufSize int
	connQueue  *algs.Queue
}

func (ql *QUICListener) Close() error {
	return ql.listener.Close()
}

func (ql *QUICListener) Addr() net.Addr {
	return ql.listener.Addr()
}

func (ql *QUICListener) Init(args map[string]interface{}) (err error) {
	return nil
}

func (ql *QUICListener) Listen(address Addr) (err error) {
	return nil
}

func (ql *QUICListener) AcceptFullConn() (conn Conn, err error) {
	session, err := ql.listener.Accept(ql.ctx)
	if err != nil {
		return nil, err
	}
	s, err := session.AcceptStream(ql.ctx)
	return newQuicConnection(s, session.LocalAddr().String(), session.RemoteAddr().String())
}

func (ql *QUICListener) AcceptWithTimeout(timeout time.Duration) (conn Conn, err error) {
	timer := time.NewTimer(timeout)
	c := make(chan interface{})
	go func() {
		conn, err = ql.AcceptFullConn()
		c <- nil
	}()
	select {
	case <-c:
		return
	case <-timer.C:
		return nil, errors.New("accept quic conn timeout")
	}
}

func (ql *QUICListener) Accept() (Conn, error) {
	return ql.AcceptFullConn()
}

func newQuicConnection(stream quic.Stream, laddr, raddr string) (*QUICConnection, error) {
	return &QUICConnection{
		Stream: stream,
		id:     newConnID(),
		lAddr:  &XAddr{QUIC, laddr},
		rAddr:  &XAddr{QUIC, raddr},
	}, nil
}

type QUICConnection struct {
	quic.Stream
	id    int64
	lAddr net.Addr
	rAddr net.Addr
}

func (qc *QUICConnection) ID() int64 {
	return qc.id
}

func (qc *QUICConnection) LocalAddr() net.Addr {
	return qc.lAddr
}

func (qc *QUICConnection) RemoteAddr() net.Addr {
	return qc.rAddr
}
func (qc *QUICConnection) SupportMux() bool {
	return true
}
