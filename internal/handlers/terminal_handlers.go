package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/auth"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/rbac"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Web SSH terminal (Settings > Terminal), ported from jniltinho/ad-manager: an SSH shell on a PTY (xterm-256color)
// carried over a WebSocket. Off unless terminal.enabled; superadmins only.
const (
	maxTerminals  = 4                // per admin
	termFrameMax  = 64 << 10         // largest WebSocket frame either way
	termPing      = 30 * time.Second // keeps reverse proxies from dropping an idle terminal
	termWriteWait = 30 * time.Second // a browser that stops reading
	termSizeMax   = 1000             // columns and rows
)

var (
	termHelloWait = 30 * time.Second // the first message must arrive within it
	termStartWait = 20 * time.Second // TCP connect, SSH handshake, authentication, PTY and shell
)

// Terminal error codes: the only error text that reaches the browser and the logs (the SSH library's text may carry
// what the server sent).
const (
	codeAuth        = "auth_failed"
	codeHostChanged = "hostkey_changed"
	codeHostUnknown = "hostkey_unknown"
	codeUnreachable = "unreachable"
	codeTimeout     = "timeout"
	codeClosed      = "closed"
	codeIdle        = "idle"
	codeInternal    = "internal"
	codeBadRequest  = "bad_request"
	codeForbidden   = "forbidden"
	codeBusy        = "busy"
)

var (
	termMu   sync.Mutex
	termOpen = map[string]int{} // open terminals per admin
)

// termHello is the WebSocket's first message: the panel token (a browser WebSocket cannot send headers) and the SSH
// credential, which live only until the handshake.
type termHello struct {
	Token      string `json:"token"`
	User       string `json:"user"`
	Password   string `json:"password"`
	Key        string `json:"key"`
	Passphrase string `json:"passphrase"`
	Cols       int    `json:"cols"`
	Rows       int    `json:"rows"`
	Trust      string `json:"trust"` // the fingerprint the operator confirmed
}

type termFrame struct {
	typ  byte
	data []byte
}

// frameCodec keeps the frame type (text = control message, binary = terminal bytes), which websocket.Message drops.
var frameCodec = websocket.Codec{
	Marshal: func(v any) ([]byte, byte, error) { f := v.(termFrame); return f.data, f.typ, nil },
	Unmarshal: func(data []byte, typ byte, v any) error {
		f := v.(*termFrame)
		f.typ, f.data = typ, data
		return nil
	},
}

type codeError struct{ code, fingerprint, keyType string }

func (e *codeError) Error() string { return e.code }

func terminalTarget() string {
	return net.JoinHostPort(viper.GetString("terminal.host"), strconv.Itoa(viper.GetInt("terminal.port")))
}

func terminalIdle() time.Duration {
	if d := viper.GetDuration("terminal.idle_timeout"); d > 0 {
		return d
	}
	return 30 * time.Minute
}

// terminalAllowed: an admin token of a superadmin who, with RBAC on, also holds settings:write.
func terminalAllowed(cl *auth.Claims) bool {
	return cl != nil && cl.Type == "admin" && cl.Superadmin &&
		(!viper.GetBool("rbac.enabled") || rbac.HasPermission(cl.Permissions, rbac.PermSettingsWrite))
}

// sameOrigin reports whether the request's Origin is the panel's own (host and port equal to Host).
func sameOrigin(r *http.Request) bool {
	o := r.Header.Get("Origin")
	if o == "" || o == "null" {
		return false
	}
	u, err := url.Parse(o)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != "" && strings.EqualFold(u.Host, r.Host)
}

// TerminalInfoV1 godoc
// @Summary      Web terminal status
// @Description  Whether the SSH terminal is enabled, its target and the pinned host key fingerprint. Superadmins only.
// @Tags         System
// @Produce      json
// @Success      200 {object} dto.TerminalInfoResponse
// @Failure      401 {object} dto.APIResponse
// @Failure      403 {object} dto.APIResponse
// @Router       /system/terminal [get]
// @Security     BearerAuth
func (h *Handler) TerminalInfoV1(c *echo.Context) error {
	if !terminalAllowed(middleware.GetJWTClaims(c)) {
		return dto.WriteError(c, dto.ErrCodeForbidden, "superadmins only")
	}
	info := dto.TerminalInfoResponse{Enabled: viper.GetBool("terminal.enabled"), Target: terminalTarget(),
		IdleMinutes: int(terminalIdle().Minutes())}
	var k models.TerminalHostKey
	if db := h.requestDB(c); db != nil && db.Where("host = ?", info.Target).First(&k).Error == nil {
		if pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(k.Key)); err == nil {
			info.HostKey = pk.Type() + " " + ssh.FingerprintSHA256(pk)
		}
	}
	return dto.WriteSuccess(c, info)
}

// ResetTerminalHostKeyV1 godoc
// @Summary      Forget the pinned SSH host key
// @Description  The next terminal connection asks the operator to trust the host key again. Superadmins only.
// @Tags         System
// @Produce      json
// @Success      200 {object} dto.APIResponse
// @Failure      401 {object} dto.APIResponse
// @Failure      403 {object} dto.APIResponse
// @Router       /system/terminal/hostkey [delete]
// @Security     BearerAuth
func (h *Handler) ResetTerminalHostKeyV1(c *echo.Context) error {
	cl := middleware.GetJWTClaims(c)
	if !terminalAllowed(cl) {
		return dto.WriteError(c, dto.ErrCodeForbidden, "superadmins only")
	}
	if h.DB == nil {
		return dto.WriteError(c, dto.ErrCodeInternal, "database unavailable")
	}
	target := terminalTarget()
	if err := h.requestDB(c).Where("host = ?", target).Delete(&models.TerminalHostKey{}).Error; err != nil {
		return dto.WriteError(c, dto.ErrCodeInternal, "could not clear the host key")
	}
	_ = utils.LogAction(h.DB, cl.Username, c.RealIP(), "", "terminal_hostkey_reset", target)
	return dto.WriteSuccess(c, map[string]string{"host": target})
}

// TerminalWS handles GET /api/v1/system/terminal/ws: a WebSocket carrying an SSH shell to terminal.host.
// Authentication is the first message's token (checked before anything is dialed), plus a same-origin check.
func (h *Handler) TerminalWS(c *echo.Context) error {
	if !viper.GetBool("terminal.enabled") {
		return dto.WriteError(c, dto.ErrCodeNotFound, "the web terminal is disabled (terminal.enabled)")
	}
	if !sameOrigin(c.Request()) {
		return dto.WriteError(c, dto.ErrCodeForbidden, "cross-origin request refused")
	}
	ip := c.RealIP()
	ctx, cancel := context.WithCancel(context.Background()) // the request's context ends with the hijack
	defer cancel()
	ws := websocket.Server{
		Handshake: func(*websocket.Config, *http.Request) error { return nil }, // Origin checked above
		Handler:   func(ws *websocket.Conn) { h.runTerminal(ctx, ws, ip) },
	}
	ws.ServeHTTP(c.Response(), c.Request())
	return nil
}

func (h *Handler) runTerminal(ctx context.Context, ws *websocket.Conn, ip string) {
	defer ws.Close()
	ws.MaxPayloadBytes = termFrameMax
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// a past deadline first: Close waits for the writer, which a browser that stopped reading would hold
	var dmu sync.Mutex
	stop := context.AfterFunc(ctx, func() {
		dmu.Lock()
		_ = ws.SetDeadline(time.Now())
		dmu.Unlock()
		ws.Close()
	})
	defer stop()
	send := func(typ byte, data []byte) error {
		dmu.Lock()
		if err := ctx.Err(); err != nil {
			dmu.Unlock()
			return err
		}
		_ = ws.SetWriteDeadline(time.Now().Add(termWriteWait))
		dmu.Unlock()
		return frameCodec.Send(ws, termFrame{typ, data})
	}
	control := func(v map[string]any) { b, _ := json.Marshal(v); _ = send(websocket.TextFrame, b) }

	_ = ws.SetDeadline(time.Now().Add(termHelloWait))
	var f termFrame
	var in termHello
	if err := frameCodec.Receive(ws, &f); err != nil {
		return // silent client, oversized frame or closed: nothing dialed
	}
	if f.typ != websocket.TextFrame || json.Unmarshal(f.data, &in) != nil {
		control(map[string]any{"type": "error", "code": codeBadRequest})
		return
	}
	claims, _ := auth.ValidateTokenOfType(in.Token, "admin")
	if !terminalAllowed(claims) {
		control(map[string]any{"type": "error", "code": codeForbidden})
		return
	}
	if strings.TrimSpace(in.User) == "" || in.Cols < 1 || in.Cols > termSizeMax || in.Rows < 1 || in.Rows > termSizeMax ||
		(in.Password == "" && in.Key == "") {
		control(map[string]any{"type": "error", "code": codeBadRequest})
		return
	}
	admin := strings.ToLower(claims.Username)
	termMu.Lock()
	if termOpen[admin] >= maxTerminals {
		termMu.Unlock()
		control(map[string]any{"type": "error", "code": codeBusy})
		return
	}
	termOpen[admin]++
	termMu.Unlock()
	defer func() {
		termMu.Lock()
		if termOpen[admin]--; termOpen[admin] <= 0 {
			delete(termOpen, admin)
		}
		termMu.Unlock()
	}()
	_ = ws.SetDeadline(time.Time{})

	target := terminalTarget()
	logTarget := in.User + "@" + target
	client, sh, stdin, stdout, err := h.startShell(ctx, target, &in, func() {
		_ = utils.LogAction(h.DB, admin, ip, "", "terminal_hostkey_trust", target)
	})
	in = termHello{} // drop the credential and the token
	if err != nil {
		var ce *codeError
		if !errors.As(err, &ce) {
			ce = &codeError{code: codeInternal}
		}
		if ce.code == codeHostUnknown {
			control(map[string]any{"type": "hostkey", "fingerprint": ce.fingerprint, "key_type": ce.keyType})
			return // nothing was sent to the server: not an attempt
		}
		slog.Warn("terminal open failed", "admin", admin, "target", target, "code", ce.code)
		_ = utils.LogAction(h.DB, admin, ip, "", "terminal_open_failed", logTarget+" "+ce.code)
		control(map[string]any{"type": "error", "code": ce.code})
		return
	}
	defer client.Close()
	_ = utils.LogAction(h.DB, admin, ip, "", "terminal_open", logTarget)
	start := time.Now()
	defer func() {
		_ = utils.LogAction(h.DB, admin, ip, "", "terminal_close", logTarget+" "+time.Since(start).Round(time.Second).String())
	}()
	control(map[string]any{"type": "ready"})

	var amu sync.Mutex
	last := time.Now() // last keystroke: the idle timeout counts from it
	ended := make(chan struct{}, 2)
	go func() { // output: binary frames
		buf := make([]byte, 32<<10)
		for {
			n, err := stdout.Read(buf)
			if n > 0 && send(websocket.BinaryFrame, buf[:n]) != nil {
				break
			}
			if err != nil {
				break
			}
		}
		ended <- struct{}{}
	}()
	go func() { // input: binary frames are keystrokes, text frames control messages
		for {
			var f termFrame
			if frameCodec.Receive(ws, &f) != nil {
				break
			}
			if f.typ == websocket.BinaryFrame {
				amu.Lock()
				last = time.Now()
				amu.Unlock()
				if _, err := stdin.Write(f.data); err != nil {
					break
				}
				continue
			}
			var r struct {
				Type       string `json:"type"`
				Cols, Rows int
			}
			if json.Unmarshal(f.data, &r) == nil && r.Type == "resize" && r.Cols >= 1 && r.Cols <= termSizeMax &&
				r.Rows >= 1 && r.Rows <= termSizeMax {
				_ = sh.WindowChange(r.Rows, r.Cols)
			}
		}
		ended <- struct{}{}
	}()
	ping, check := time.NewTicker(termPing), time.NewTicker(15*time.Second)
	defer ping.Stop()
	defer check.Stop()
	idle := terminalIdle()
	for {
		select {
		case <-ended:
			control(map[string]any{"type": "error", "code": codeClosed})
			return
		case <-ctx.Done():
			return
		case <-ping.C:
			_ = send(websocket.PingFrame, nil)
		case <-check.C:
			amu.Lock()
			quiet := time.Since(last)
			amu.Unlock()
			if quiet > idle {
				control(map[string]any{"type": "error", "code": codeIdle})
				return
			}
		}
	}
}

// startShell dials target, checks the pinned host key, authenticates and starts a shell on a PTY, all within
// termStartWait.
func (h *Handler) startShell(ctx context.Context, target string, in *termHello, trusted func()) (*ssh.Client, *ssh.Session, io.Writer, io.Reader, error) {
	if h.DB == nil {
		return nil, nil, nil, nil, &codeError{code: codeInternal}
	}
	var auths []ssh.AuthMethod
	if in.Key != "" {
		var signer ssh.Signer
		var err error
		if in.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(in.Key), []byte(in.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(in.Key))
		}
		if err != nil {
			return nil, nil, nil, nil, &codeError{code: codeAuth}
		}
		auths = append(auths, ssh.PublicKeys(signer))
	}
	if pw := in.Password; pw != "" {
		auths = append(auths, ssh.Password(pw), ssh.KeyboardInteractive(func(_, _ string, q []string, _ []bool) ([]string, error) {
			a := make([]string, len(q))
			for i := range a {
				a[i] = pw
			}
			return a, nil
		}))
	}
	db := h.DB.WithContext(ctx)
	var pinned models.TerminalHostKey
	_ = db.Where("host = ?", target).First(&pinned).Error
	trust := in.Trust
	hostKey := func(_ string, _ net.Addr, key ssh.PublicKey) error {
		return checkHostKey(db, target, pinned.Key, trust, key, trusted)
	}
	cfg := &ssh.ClientConfig{User: in.User, Auth: auths, HostKeyCallback: hostKey,
		HostKeyAlgorithms: hostKeyAlgos(pinned.Key), Timeout: termStartWait}

	ctx, cancel := context.WithTimeout(ctx, termStartWait)
	defer cancel()
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", target)
	if err != nil {
		return nil, nil, nil, nil, startError(ctx, err)
	}
	stop := context.AfterFunc(ctx, func() { conn.Close() }) // a server that stalls after the TCP connect
	cc, chans, reqs, err := ssh.NewClientConn(conn, target, cfg)
	if err != nil {
		stop()
		conn.Close()
		return nil, nil, nil, nil, startError(ctx, err)
	}
	client := ssh.NewClient(cc, chans, reqs)
	sh, stdin, stdout, err := shell(client, in.Cols, in.Rows)
	if !stop() || err != nil { // the deadline fired (the connection is closed) or the shell failed
		client.Close()
		if err == nil {
			err = context.DeadlineExceeded
		}
		return nil, nil, nil, nil, startError(ctx, err)
	}
	return client, sh, stdin, stdout, nil
}

// checkHostKey accepts the pinned key; with none pinned it pins the key only when the operator confirmed its
// fingerprint (trust), otherwise it reports it for confirmation. A different key than the pinned one is refused.
func checkHostKey(db *gorm.DB, target, pinned, trust string, key ssh.PublicKey, trusted func()) error {
	line := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
	fp := ssh.FingerprintSHA256(key)
	if pinned != "" {
		if pinned == line {
			return nil
		}
		return &codeError{code: codeHostChanged}
	}
	if trust != fp {
		return &codeError{code: codeHostUnknown, fingerprint: fp, keyType: key.Type()}
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.TerminalHostKey{Host: target, Key: line}).Error; err != nil {
		return &codeError{code: codeInternal}
	}
	// read back instead of trusting RowsAffected (MySQL's clientFoundRows changes it): if another connection pinned a
	// different key meanwhile, that one wins and this key is refused
	var now models.TerminalHostKey
	if err := db.Where("host = ?", target).First(&now).Error; err != nil {
		return &codeError{code: codeInternal}
	}
	if now.Key != line {
		return &codeError{code: codeHostChanged}
	}
	trusted()
	return nil
}

// hostKeyAlgos asks for the pinned key's own type (another type the server gained later is not a "changed" key) or,
// with none pinned, prefers Ed25519, the key operators usually check.
func hostKeyAlgos(pinned string) []string {
	if pinned == "" {
		return []string{ssh.KeyAlgoED25519, ssh.KeyAlgoECDSA256, ssh.KeyAlgoECDSA384, ssh.KeyAlgoECDSA521,
			ssh.KeyAlgoRSASHA512, ssh.KeyAlgoRSASHA256}
	}
	k, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pinned))
	switch {
	case err != nil:
		return nil // the library's defaults; the callback still compares the whole key
	case k.Type() == ssh.KeyAlgoRSA:
		return []string{ssh.KeyAlgoRSASHA512, ssh.KeyAlgoRSASHA256}
	}
	return []string{k.Type()}
}

// shell opens a session with an xterm-256color PTY and a login shell; stderr shares the PTY's output.
func shell(client *ssh.Client, cols, rows int) (*ssh.Session, io.Writer, io.Reader, error) {
	sh, err := client.NewSession()
	if err != nil {
		return nil, nil, nil, err
	}
	stdin, err := sh.StdinPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stdout, err := sh.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	if err := sh.RequestPty("xterm-256color", rows, cols, ssh.TerminalModes{ssh.ECHO: 1}); err != nil {
		return nil, nil, nil, err
	}
	if err := sh.Shell(); err != nil {
		return nil, nil, nil, err
	}
	return sh, stdin, stdout, nil
}

// startError reduces a start failure to its code.
func startError(ctx context.Context, err error) error {
	var ce *codeError
	var ne net.Error
	switch {
	case errors.As(err, &ce):
		return ce
	case ctx.Err() != nil || errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &ne) && ne.Timeout()):
		return &codeError{code: codeTimeout}
	case strings.Contains(err.Error(), "unable to authenticate"):
		return &codeError{code: codeAuth}
	case errors.As(err, new(*net.OpError)):
		return &codeError{code: codeUnreachable}
	}
	return &codeError{code: codeClosed}
}
