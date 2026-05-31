# Spec — migration-vue3-jwt

**Objetivo**: Transformar o go-postfixadmin de SSR (Echo + Go templates + jQuery) em API REST com JWT + Vue 3 SPA embarcada no binário Go.

**Status**: ✅ Concluído (com desvio: Quasar substituído por Tailwind CSS + Lucide)

---

## Contexto e Motivação

Stack anterior: Echo v5 + Go templates + Tailwind (standalone) + jQuery + DataTables + sessões cookie.

Razões da migração:
- Templates + jQuery difíceis de evoluir e manter
- Sessões cookie vulneráveis a certos ataques; JWT é padrão moderno para SPAs e habilita clientes futuros
- SPA com hot-reload e contratos REST tipados melhoram DX dramaticamente
- Preservar single-binary (crítico para empacotamento deb/rpm/Docker)

---

## Arquitetura Final

```
Single Go Binary (//go:embed)
  ├── Echo Server
  │    ├── /api/v1/*  → JWT middleware → Handlers → Repositories
  │    └── /*         → SPA static + index.html fallback (history mode)
  └── Embedded FS: web/dist + locales + web/static
```

- Frontend: `frontend/` (Vue 3 + Vite + Tailwind CSS v4 + Lucide) → build para `web/dist/`
- Backend serve a SPA como fallback para todas as rotas não-API
- CLI, transport TCP, maillog reader: **inalterados**

---

## JWT Strategy

**Access Token** (15-30 min, Bearer header):
```json
{
  "sub": "user@example.com",
  "type": "admin" | "mailbox",
  "superadmin": true | false,
  "domains": ["example.com"] | ["ALL"],
  "iat": 1234567890,
  "exp": 1234569690,
  "iss": "go-postfixadmin"
}
```

**Refresh Token** (7 dias, httpOnly cookie):
- Stateless JWT com claims mínimos (`sub`, `jti`, `exp`)
- Path: `/api/v1/auth/refresh`, SameSite=Lax, Secure em produção
- Rotação a cada uso (novo `jti` emitido)
- **Sem tabela de revogação em v1** — matching da simplicidade das sessões anteriores

**Password change**: força re-login (response com `force_relogin: true`), cliente limpa store + cookies.

---

## API Response Envelope

```json
// Sucesso
{ "success": true, "data": { ... } }

// Erro (com HTTP status correto)
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR|UNAUTHORIZED|FORBIDDEN|NOT_FOUND|CONFLICT|RATE_LIMITED",
    "message": "mensagem legível",
    "details": [{ "field": "password", "message": "..." }]
  }
}
```

---

## SPA Embedding (Go)

```go
//go:embed web/dist locales web/static
var embeddedFiles embed.FS
```

Handler de fallback SPA (`internal/server/spa.go`):
```go
func SPAFileServer(embedded embed.FS, distRoot string) echo.HandlerFunc {
    distFS, _ := fs.Sub(embedded, distRoot)
    fileServer := http.FileServer(http.FS(distFS))
    return func(c echo.Context) error {
        path := strings.TrimPrefix(c.Request().URL.Path, "/")
        // 1. Arquivo existe → serve direto
        if f, err := distFS.Open(path); err == nil {
            f.Close()
            return echo.WrapHandler(http.StripPrefix("/", fileServer))(c)
        }
        // 2. Rota SPA (sem extensão) → index.html
        if c.Request().Method == http.MethodGet && !strings.Contains(path, ".") {
            if index, err := distFS.Open("index.html"); err == nil {
                defer index.Close()
                return c.Stream(http.StatusOK, "text/html; charset=utf-8", index)
            }
        }
        // 3. 404 para assets inválidos
        return echo.WrapHandler(fileServer)(c)
    }
}
```

Registrado **por último**, após todas as rotas `/api/v1/*` e `/static/*`.

---

## Dev Local (hot-reload)

Terminal 1 (backend):
```bash
go run main.go server   # :8080
```

Terminal 2 (frontend):
```bash
cd frontend && npm run dev   # Vite em :5173 com proxy para :8080
```

`vite.config.ts` proxy:
```ts
proxy: {
  '/api': { target: 'http://localhost:8080', changeOrigin: true },
  '/lang': { target: 'http://localhost:8080', changeOrigin: true }
}
```

CORS dev-only no Go (habilitado quando `debug=true`):
- `AllowOrigins: ["http://localhost:5173"]`
- `AllowCredentials: true` (necessário para cookies httpOnly)
- Refresh cookie com `Secure: false` em localhost

---

## Segurança

- JWT: expiry curto + rotação de refresh + re-login obrigatório após troca de senha
- Domain scoping: nunca confiar no cliente — sempre re-validar no DB ou nas claims assinadas
- Login rate limiting: por IP + por username (429 + `Retry-After`)
- CSRF: `SameSite=Lax` + checagem de `Origin`/`Referer` no refresh handler
- XSS: Vue escaping automático, sem `innerHTML` de dados não confiáveis
- Access token em memória (não `localStorage`); refresh em httpOnly cookie
- `LogAction` chamado em todas as mutações (preservado dos handlers originais)
- TLS recomendado em produção (`ssl_enable=true`)

---

## i18n

- **Backend** (erros de API, emails, logs): mantém `.po` + gotext (`internal/i18n/`)
- **Frontend** (labels, toasts, validações): vue-i18n com JSON (`frontend/src/locales/`)
- Mesmas chaves dos `.po` onde há overlap
- Language cookie respeitado por ambos
- `internal/i18n.GetPreferredLang()` centraliza lógica duplicada anterior

---

## Decisões-Chave

1. **JWT sobre sessões** — padrão moderno, habilita clientes futuros, refresh httpOnly resiste a XSS
2. **Single-binary embed** — não-negociável para manter deb/rpm/Docker simples
3. **Tailwind + Lucide ao invés de Quasar** — fidelidade visual neo-brutalist com menos conflito de CSS
4. **Rollout incremental** — dual-mode (flag `features.spa`) até cutover final
5. **Sem mudança de schema DB** — lógica de negócio e modelos já corretos
6. **Evoluir handlers existentes** — scoping, password policy, Sieve sync já funcionam
7. **v1 sem tabela de revogação** — matching da simplicidade das sessões anteriores
