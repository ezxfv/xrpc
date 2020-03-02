package net

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"time"
)

const (
	DefaultTimeout = time.Second
)

func WrapTLSClient(conn net.Conn, tlsConfig *tls.Config, timeout time.Duration) (net.Conn, error) {
	var err error
	var tlsConn *tls.Conn

	if timeout <= 0 {
		timeout = DefaultTimeout // default timeout
	}

	conn.SetDeadline(time.Now().Add(timeout))
	defer conn.SetDeadline(time.Time{})

	tlsConn = tls.Client(conn, tlsConfig)

	// Otherwise perform handshake, but don't verify the domain
	//
	// The following is lightly-modified from the doFullHandshake
	// method in https://golang.org/src/crypto/tls/handshake_client.go
	if err = tlsConn.Handshake(); err != nil {
		tlsConn.Close()
		return nil, err
	}

	// If crypto/tls is doing verification, there's no need to do our own.
	if tlsConfig.InsecureSkipVerify == false {
		return tlsConn, nil
	}

	// Similarly if we use host's CA, we can do full handshake
	if tlsConfig.RootCAs == nil {
		return tlsConn, nil
	}

	opts := x509.VerifyOptions{
		Roots:         tlsConfig.RootCAs,
		CurrentTime:   time.Now(),
		DNSName:       "",
		Intermediates: x509.NewCertPool(),
	}

	certs := tlsConn.ConnectionState().PeerCertificates
	for i, cert := range certs {
		if i == 0 {
			continue
		}
		opts.Intermediates.AddCert(cert)
	}

	_, err = certs[0].Verify(opts)
	if err != nil {
		tlsConn.Close()
		return nil, err
	}

	return tlsConn, err
}
