package email

import (
	"errors"
	"net/smtp"
	"strings"
)

var (
	errSMTPAuthRequiresTLS      = errors.New("smtp authentication requires TLS")
	errSMTPAuthHostMismatch     = errors.New("smtp authentication host mismatch")
	errSMTPAuthMechanismMissing = errors.New("smtp server does not advertise a supported authentication mechanism")
	errSMTPAuthChallengeInvalid = errors.New("smtp authentication challenge is unsupported")
)

// NewSMTPAuth negotiates only mechanisms advertised by the SMTP server.
// Credentials are never returned before TLS and hostname checks pass.
func NewSMTPAuth(username, password, host string) smtp.Auth {
	return &negotiatedSMTPAuth{
		username: strings.TrimSpace(username),
		password: password,
		host:     strings.TrimSpace(host),
	}
}

type negotiatedSMTPAuth struct {
	username string
	password string
	host     string
}

func (a *negotiatedSMTPAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS {
		return "", nil, errSMTPAuthRequiresTLS
	}
	if !strings.EqualFold(strings.TrimSpace(server.Name), a.host) {
		return "", nil, errSMTPAuthHostMismatch
	}
	mechanisms := make(map[string]struct{}, len(server.Auth))
	for _, mechanism := range server.Auth {
		mechanisms[strings.ToUpper(strings.TrimSpace(mechanism))] = struct{}{}
	}
	if _, ok := mechanisms["LOGIN"]; ok {
		return "LOGIN", nil, nil
	}
	if _, ok := mechanisms["PLAIN"]; ok {
		return "PLAIN", []byte("\x00" + a.username + "\x00" + a.password), nil
	}
	return "", nil, errSMTPAuthMechanismMissing
}

func (a *negotiatedSMTPAuth) Next(challenge []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	prompt := strings.ToLower(strings.TrimSpace(string(challenge)))
	switch {
	case strings.Contains(prompt, "username"):
		return []byte(a.username), nil
	case strings.Contains(prompt, "password"):
		return []byte(a.password), nil
	default:
		return nil, errSMTPAuthChallengeInvalid
	}
}
