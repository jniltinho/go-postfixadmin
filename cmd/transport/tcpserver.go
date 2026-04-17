package transport

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
	zlog "github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// TCPServer serves Postfix transport map queries over TCP.
// Postfix main.cf: transport_maps = tcp:127.0.0.1:12221
type TCPServer struct {
	address       string
	db            *gorm.DB
	cache         *cache.Cache
	hostname      string
	localDelivery string
	delivery      string
	debug         bool
}

// NewTCPServer creates and returns a new TCPServer.
func NewTCPServer(address string, db *gorm.DB, cacheDur time.Duration, hostname, localDelivery, delivery string, debug bool) *TCPServer {
	return &TCPServer{
		address:       address,
		db:            db,
		cache:         cache.New(cacheDur, 5*time.Minute),
		hostname:      hostname,
		localDelivery: localDelivery,
		delivery:      delivery,
		debug:         debug,
	}
}

// Start begins accepting connections. Blocks until a fatal error occurs.
func (s *TCPServer) Start() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	defer listener.Close()

	if s.debug {
		zlog.Info().Msg("Debug mode activated")
	}
	zlog.Info().Msgf("Starting TCP server on %s...", s.address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			zlog.Error().Msgf("Error accepting connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *TCPServer) logEvent(source, subject, domain string, isEmail bool, ret string) {
	if !s.debug {
		return
	}
	ev := zlog.Info().Str("cmd", "get "+subject)
	if source == "DB" {
		ev.Str("domain", domain)
	} else {
		if isEmail {
			ev.Str("email", subject)
		} else {
			ev.Str("domain", subject)
		}
	}
	ev.Str("ret", ret).Msg(source)
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if strings.ToLower(parts[0]) != "get" || len(parts) < 2 {
			continue
		}

		rawSubject := parts[1]
		subject, err := url.PathUnescape(rawSubject)
		if err != nil {
			s.logEvent("EXTERNAL", rawSubject, "", false, "400 INVALID ENCODING")
			s.reply(conn, "400 INVALID ENCODING")
			continue
		}
		subject = strings.ToLower(subject)

		if subject == "*" {
			s.logEvent("EXTERNAL", subject, "", false, "500 not found")
			s.reply(conn, "500 not found")
			continue
		}

		if strings.ContainsAny(subject, "'\"") {
			s.logEvent("EXTERNAL", subject, "", false, "400 invalid caracter")
			s.reply(conn, "400 invalid caracter")
			continue
		}

		domain := subject
		isEmail := false
		if idx := strings.LastIndex(subject, "@"); idx != -1 {
			domain = subject[idx+1:]
			isEmail = true
		}

		if cached, found := s.cache.Get(subject); found {
			dest := cached.(string)
			if dest == "" {
				s.logEvent("CACHE", subject, domain, isEmail, "500 not found")
				s.reply(conn, "500 not found")
			} else {
				s.logEvent("CACHE", subject, domain, isEmail, "200 "+dest)
				s.reply(conn, "200 "+url.PathEscape(dest))
			}
			continue
		}

		email := ""
		if isEmail {
			email = subject
		}

		dest, dbErr := GetTransport(s.db, email, domain, s.hostname, s.localDelivery, s.delivery)
		if dbErr != nil {
			zlog.Error().Msgf("DB Error: %v", dbErr)
			s.logEvent("DB", subject, domain, isEmail, "400 internal error")
			s.reply(conn, "400 internal error")
			continue
		}

		if dest == "" {
			s.logEvent("DB", subject, domain, isEmail, "500 not found")
			s.cache.SetDefault(subject, "")
			s.reply(conn, "500 not found")
		} else {
			s.logEvent("DB", subject, domain, isEmail, "200 "+dest)
			s.cache.SetDefault(subject, dest)
			s.reply(conn, "200 "+url.PathEscape(dest))
		}
	}
}

func (s *TCPServer) reply(conn net.Conn, msg string) {
	fmt.Fprintf(conn, "%s\n", msg)
}
