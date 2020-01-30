package net

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/edenzhong7/xrpc/pkg/log"

	"golang.org/x/crypto/ssh"
)

func init() {
	RegisterDialer(SSH, sshDialer)
	RegisterListenerBuilder(SSH, newSSHListener)
}

var (
	sshDialer Dialer = func(ctx context.Context, addr string) (conn Conn, err error) {
		host, ok := ctx.Value("hostname").(string)
		if !ok {
			return nil, errors.New("ssh: get hostname from ctx failed")
		}
		file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		var hostKey ssh.PublicKey
		for scanner.Scan() {
			fields := strings.Split(scanner.Text(), " ")
			if len(fields) != 3 {
				continue
			}
			if strings.Contains(fields[0], host) {
				var err error
				hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
				if err != nil {
					log.Fatalf("error parsing %q: %v", fields[2], err)
				}
				break
			}
		}

		if hostKey == nil {
			log.Fatalf("no hostkey for %s", host)
		}

		config := ssh.ClientConfig{
			User:            os.Getenv("USER"),
			HostKeyCallback: ssh.FixedHostKey(hostKey),
		}

		conn, err = tcpDialer(ctx, addr)
		sconn, chans, reqs, err := ssh.NewClientConn(conn, addr, &config)
		if err != nil {
			conn.Close()
			return nil, err
		}
		client := ssh.NewClient(sconn, chans, reqs)
		channel, reqs, err := client.OpenChannel(SSH, nil)
		return &sshConn{
			Conn:    conn,
			channel: channel,
		}, nil
	}
)

func newSSHListener(ctx context.Context, addr string) (lis Listener, err error) {
	config, ok := ctx.Value("ssh_config").(*SSHConfig)
	if !ok {
		return nil, errors.New("can't read ssh config from ctx")
	}
	ln, err := newTCPListener(ctx, addr)
	if err != nil {
		return nil, err
	}

	if config == nil {
		config = &SSHConfig{}
	}

	sshConfig := &ssh.ServerConfig{}
	//sshConfig.PasswordCallback = defaultSSHPasswordCallback(config.Authenticator)
	if config.Authenticator == nil {
		sshConfig.NoClientAuth = true
	}
	tlsConfig := config.TLSConfig

	signer, err := ssh.NewSignerFromKey(tlsConfig.Certificates[0].PrivateKey)
	if err != nil {
		ln.Close()
		return nil, err

	}
	sshConfig.AddHostKey(signer)

	l := &sshListener{
		Listener: ln,
		config:   sshConfig,
		connChan: make(chan Conn, 1024),
		errChan:  make(chan error, 1),
	}

	go l.listenLoop()

	return l, nil
}

type SSHConfig struct {
	Authenticator Authenticator
	TLSConfig     *tls.Config
}

type sshConn struct {
	Conn

	channel ssh.Channel
}

func (s sshConn) Close() error {
	return s.channel.Close()
}

func (s sshConn) Read(b []byte) (n int, err error) {
	return s.channel.Read(b)
}

func (s sshConn) Write(b []byte) (n int, err error) {
	return s.channel.Write(b)
}

type sshListener struct {
	Listener
	config   *ssh.ServerConfig
	connChan chan Conn
	errChan  chan error
}

func (l *sshListener) listenLoop() {
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			log.Debug("[ssh] accept:", err)
			l.errChan <- err
			close(l.errChan)
			return
		}
		go l.serveConn(conn)
	}
}

func (l *sshListener) serveConn(conn Conn) {
	sc, chans, reqs, err := ssh.NewServerConn(conn, l.config)
	if err != nil {
		log.Debugf("[ssh] %s -> %s : %s", conn.RemoteAddr(), conn.LocalAddr(), err)
		conn.Close()
		return
	}
	defer sc.Close()

	go ssh.DiscardRequests(reqs)
	go func() {
		for newChannel := range chans {
			// Check the type of channel
			t := newChannel.ChannelType()
			switch t {
			case SSH:
				channel, requests, err := newChannel.Accept()
				if err != nil {
					log.Debug("[ssh] Could not accept channel:", err)
					continue
				}
				go ssh.DiscardRequests(requests)
				cc := &sshConn{Conn: conn, channel: channel}
				select {
				case l.connChan <- cc:
				default:
					cc.Close()
					log.Debugf("[ssh] %s - %s: connection queue is full", conn.RemoteAddr(), l.Addr())
				}

			default:
				log.Debug("[ssh] Unknown channel type:", t)
				newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			}
		}
	}()

	log.Debugf("[ssh] %s <-> %s", conn.RemoteAddr(), conn.LocalAddr())
	sc.Wait()
	log.Debugf("[ssh] %s >-< %s", conn.RemoteAddr(), conn.LocalAddr())
}

func (l *sshListener) Accept() (conn Conn, err error) {
	var ok bool
	select {
	case conn = <-l.connChan:
	case err, ok = <-l.errChan:
		if !ok {
			err = errors.New("accpet on closed listener")
		}
	}
	return
}
