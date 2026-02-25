package socks5

import (
	"fmt"
	"io"
	"net"
)

const (
	NoAuth          = uint8(0)
	noAcceptable    = uint8(255)
	UserPassAuth    = uint8(2)
	userAuthVersion = uint8(1)
	authSuccess     = uint8(0)
	authFailure     = uint8(1)
)

var (
	UserAuthFailed  = fmt.Errorf("User authentication failed")
	NoSupportedAuth = fmt.Errorf("No supported authentication mechanism")
)

// authFailureError captures structured authentication failure details so callers
// can log consistent entries (for example, fail2ban-friendly log lines).
type authFailureError struct {
	Method     uint8
	Username   string
	Reason     string
	RemoteIP   string
	RemoteAddr string
	Err        error
}

func (e *authFailureError) Error() string {
	return fmt.Sprintf(
		"authentication failed: method=%d username=%q remote_ip=%s remote_addr=%q reason=%s: %v",
		e.Method,
		e.Username,
		e.RemoteIP,
		e.RemoteAddr,
		e.Reason,
		e.Err,
	)
}

func (e *authFailureError) Unwrap() error {
	return e.Err
}

func newAuthFailureError(writer io.Writer, method uint8, username, reason string, err error) error {
	remoteAddr, remoteIP := getRemoteAddressAndIP(writer)
	return &authFailureError{
		Method:     method,
		Username:   username,
		Reason:     reason,
		RemoteIP:   remoteIP,
		RemoteAddr: remoteAddr,
		Err:        err,
	}
}

func getRemoteAddressAndIP(writer io.Writer) (string, string) {
	remoteAddr := "unknown"
	remoteIP := "unknown"

	addrProvider, ok := writer.(interface{ RemoteAddr() net.Addr })
	if !ok {
		return remoteAddr, remoteIP
	}

	addr := addrProvider.RemoteAddr()
	if addr == nil {
		return remoteAddr, remoteIP
	}

	remoteAddr = addr.String()

	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		if parsed := net.ParseIP(remoteAddr); parsed != nil {
			return remoteAddr, parsed.String()
		}
		return remoteAddr, remoteIP
	}

	if parsed := net.ParseIP(host); parsed != nil {
		remoteIP = parsed.String()
	} else {
		remoteIP = host
	}

	return remoteAddr, remoteIP
}

// A Request encapsulates authentication state provided
// during negotiation
type AuthContext struct {
	// Provided auth method
	Method uint8
	// Payload provided during negotiation.
	// Keys depend on the used auth method.
	// For UserPassauth contains Username
	Payload map[string]string
}

type Authenticator interface {
	Authenticate(reader io.Reader, writer io.Writer) (*AuthContext, error)
	GetCode() uint8
}

// NoAuthAuthenticator is used to handle the "No Authentication" mode
type NoAuthAuthenticator struct{}

func (a NoAuthAuthenticator) GetCode() uint8 {
	return NoAuth
}

func (a NoAuthAuthenticator) Authenticate(reader io.Reader, writer io.Writer) (*AuthContext, error) {
	_, err := writer.Write([]byte{socks5Version, NoAuth})
	return &AuthContext{NoAuth, nil}, err
}

// UserPassAuthenticator is used to handle username/password based
// authentication
type UserPassAuthenticator struct {
	Credentials CredentialStore
}

func (a UserPassAuthenticator) GetCode() uint8 {
	return UserPassAuth
}

func (a UserPassAuthenticator) Authenticate(reader io.Reader, writer io.Writer) (*AuthContext, error) {
	// Tell the client to use user/pass auth
	if _, err := writer.Write([]byte{socks5Version, UserPassAuth}); err != nil {
		return nil, newAuthFailureError(writer, UserPassAuth, "", "write_auth_method_failed", err)
	}

	// Get the version and username length
	header := []byte{0, 0}
	if _, err := io.ReadAtLeast(reader, header, 2); err != nil {
		return nil, newAuthFailureError(writer, UserPassAuth, "", "read_auth_header_failed", err)
	}

	// Ensure we are compatible
	if header[0] != userAuthVersion {
		return nil, newAuthFailureError(
			writer,
			UserPassAuth,
			"",
			"unsupported_auth_version",
			fmt.Errorf("Unsupported auth version: %v", header[0]),
		)
	}

	// Get the user name
	userLen := int(header[1])
	user := make([]byte, userLen)
	if _, err := io.ReadAtLeast(reader, user, userLen); err != nil {
		return nil, newAuthFailureError(writer, UserPassAuth, "", "read_username_failed", err)
	}

	// Get the password length
	if _, err := reader.Read(header[:1]); err != nil {
		return nil, newAuthFailureError(writer, UserPassAuth, string(user), "read_password_length_failed", err)
	}

	// Get the password
	passLen := int(header[0])
	pass := make([]byte, passLen)
	if _, err := io.ReadAtLeast(reader, pass, passLen); err != nil {
		return nil, newAuthFailureError(writer, UserPassAuth, string(user), "read_password_failed", err)
	}

	// Verify the password
	if a.Credentials.Valid(string(user), string(pass)) {
		if _, err := writer.Write([]byte{userAuthVersion, authSuccess}); err != nil {
			return nil, newAuthFailureError(writer, UserPassAuth, string(user), "write_auth_success_failed", err)
		}
	} else {
		if _, err := writer.Write([]byte{userAuthVersion, authFailure}); err != nil {
			return nil, newAuthFailureError(writer, UserPassAuth, string(user), "write_auth_failure_failed", err)
		}
		return nil, newAuthFailureError(writer, UserPassAuth, string(user), "invalid_credentials", UserAuthFailed)
	}

	// Done
	return &AuthContext{UserPassAuth, map[string]string{"Username": string(user)}}, nil
}

// authenticate is used to handle connection authentication
func (s *Server) authenticate(conn io.Writer, bufConn io.Reader) (*AuthContext, error) {
	// Get the methods
	methods, err := readMethods(bufConn)
	if err != nil {
		return nil, newAuthFailureError(conn, noAcceptable, "", "read_auth_methods_failed", fmt.Errorf("Failed to get auth methods: %v", err))
	}

	// Select a usable method
	for _, method := range methods {
		cator, found := s.authMethods[method]
		if found {
			ctx, err := cator.Authenticate(bufConn, conn)
			if err != nil {
				if _, ok := err.(*authFailureError); ok {
					return nil, err
				}
				return nil, newAuthFailureError(conn, method, "", "auth_method_failed", err)
			}
			return ctx, nil
		}
	}

	// No usable method found
	return nil, newAuthFailureError(conn, noAcceptable, "", "no_supported_auth_method", noAcceptableAuth(conn))
}

// noAcceptableAuth is used to handle when we have no eligible
// authentication mechanism
func noAcceptableAuth(conn io.Writer) error {
	conn.Write([]byte{socks5Version, noAcceptable})
	return NoSupportedAuth
}

// readMethods is used to read the number of methods
// and proceeding auth methods
func readMethods(r io.Reader) ([]byte, error) {
	header := []byte{0}
	if _, err := r.Read(header); err != nil {
		return nil, err
	}

	numMethods := int(header[0])
	methods := make([]byte, numMethods)
	_, err := io.ReadAtLeast(r, methods, numMethods)
	return methods, err
}
