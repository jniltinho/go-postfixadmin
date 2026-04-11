package server

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"transport/internal/database"

	"github.com/patrickmn/go-cache"
	zlog "github.com/rs/zerolog/log"
)

type TCPServer struct {
	Address string
	DB      *database.Database
	Cache   *cache.Cache
	Debug   bool
}

func New(address string, db *database.Database, cacheDur time.Duration, debug bool) *TCPServer {
	c := cache.New(cacheDur, 5*time.Minute)
	return &TCPServer{
		Address: address,
		DB:      db,
		Cache:   c,
		Debug:   debug,
	}
}

func (s *TCPServer) Start() error {
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}
	defer listener.Close()

	if s.Debug {
		zlog.Info().Msg("Debug Line activated")
	}
	zlog.Info().Msgf("Starting TCP server on %s...", s.Address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			zlog.Error().Msgf("Error accepting connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *TCPServer) logEvent(source string, subject string, domain string, isEmail bool, ret string) {
	if !s.Debug {
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
		cmd := strings.ToLower(parts[0])
		if cmd != "get" || len(parts) < 2 {
			continue
		}

		rawSubject := parts[1]
		subject, err := url.PathUnescape(rawSubject)
		if err != nil {
			s.logEvent("EXTERNAL", rawSubject, "", false, "400 INVALID ENCODING")
			s.reply(conn, "400 INVALID ENCODING")
			continue
		}

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

		// Cache hit
		if cached, found := s.Cache.Get(subject); found {
			dest := cached.(string)
			if dest == "" {
				s.logEvent("CACHE", subject, domain, isEmail, "500 not found")
				s.reply(conn, "500 not found")
			} else {
				// Use non-encoded dest for logging to match criare-antispam/transport
				s.logEvent("CACHE", subject, domain, isEmail, "200 "+dest)
				s.reply(conn, "200 "+url.PathEscape(dest))
			}
			continue
		}

		var dest string
		var dbErr error

		// Fallback to manual perl logic or normal query
		if isEmail {
			dest, dbErr = s.DB.GetEmailTransport(subject, domain)
		} else {
			dest, dbErr = s.DB.GetDomainTransport(domain)
		}

		if dbErr != nil {
			zlog.Error().Msgf("DB Error: %v", dbErr)
			s.logEvent("DB", subject, domain, isEmail, "400 internal error")
			s.reply(conn, "400 internal error")
			continue
		}

		if dest == "" {
			s.logEvent("DB", subject, domain, isEmail, "500 not found")
			s.Cache.SetDefault(subject, "")
			s.reply(conn, "500 not found")
		} else {
			s.logEvent("DB", subject, domain, isEmail, "200 "+dest)
			s.Cache.SetDefault(subject, dest)
			s.reply(conn, "200 "+url.PathEscape(dest))
		}
	}
}

func (s *TCPServer) reply(conn net.Conn, msg string) {
	fmt.Fprintf(conn, "%s\n", msg)
}
