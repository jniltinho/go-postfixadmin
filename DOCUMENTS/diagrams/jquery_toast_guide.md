# Implementando Flash Messages e jQuery Toast num Projeto Golang

Este guia detalha quais bibliotecas, arquivos e funções você precisa importar para ter o sistema de Flash Messages do backend (Golang) se comunicando via toast.js (jQuery) no frontend.

## 1. Bibliotecas (Libs) Necessárias

### 1.1 Backend (Go)
Você precisa de uma biblioteca de sessões para conseguir armazenar as flash messages através de cookies no client (browser):
```bash
go get github.com/gorilla/sessions
```
*(Se estiver usando Echo ou Gin, as libs já possuem wrap helpers, mas a base padrão é usar sessions)*

### 1.2 Frontend (HTML/JS/CSS)
Para a versão cliente você vai precisar de:
1. **jQuery** (`jquery.min.js` v3 ou superior)
2. **jQuery Toast** (`jquery.toast.min.js` e `jquery.toast.min.css`): a lib principal.

Você pode baixá-las localmente no seu projeto (ex: `web/static/js...`) ou usar CDN.

---

## 2. Implementação no Backend (Golang)

### 2.1 Funções Base (Exemplo em um `helpers.go` ou `base_handler.go`)
Primeiro adicione uma forma de gravar o Hash no Cookie Store da sessão.

```go
package handlers

import (
	"net/http"
	"github.com/gorilla/sessions"
)

type BaseHandler struct {
	Store       *sessions.CookieStore
	SessionName string
}

// setFlash guarda a flash message na sessão de forma temporária.
func (b *BaseHandler) setFlash(r *http.Request, w http.ResponseWriter, level, message string) {
	// Pega a sessão
	sess, _ := b.Store.Get(r, b.SessionName)
	// Salva na chave o tipo da mensagem ("success", "error", etc)
	sess.AddFlash(message, level)
	sess.Save(r, w)
}

// getFlash faz a leitura e exclusão da mensagem (Flash = lê uma vez e some)
func (b *BaseHandler) getFlash(r *http.Request, w http.ResponseWriter) (string, string) {
	sess, err := b.Store.Get(r, b.SessionName)
	if err != nil {
		return "", ""
	}
	
	for _, level := range []string{"success", "error", "warning", "info"} {
		if flashes := sess.Flashes(level); len(flashes) > 0 {
			sess.Save(r, w)
			return flashes[0].(string), level 
		}
	}
	return "", ""
}
```

### 2.2 Função Auxiliar (Tratando a Primeira Letra p/ Header do Toast)
```go
func flashType(level string) string {
	if level == "success" {
		return "Success"
	}
	if level == "error" {
	    return "Error"
	}
	return "Info"
}
```

### 2.3 Enviando os dados de Flash Message para o Template
Ao invés de passar struct vazias, preencha o template capturando os possíveis Flashes criados antes.

```go
func (b *BaseHandler) BaseData(r *http.Request, w http.ResponseWriter) map[string]any {
    data := make(map[string]any)
    
	msg, level := b.getFlash(r, w)
	if msg != "" {
		data["Flash"] = msg
		data["FlashType"] = flashType(level) // Ex: "Success" ou "Error"
		data["FlashIcon"] = level            // Ex: "success" ou "error"
	}
    
    return data
}
```

### 2.4 Usando nos Handlers / Controllers
```go
func (h *Handler) MeuEndpointAcao(w http.ResponseWriter, r *http.Request) {
    // Ação efetuada com sucesso:
    h.setFlash(r, w, "success", "Registro inserido com sucesso!")
    http.Redirect(w, r, "/minha-tabela", http.StatusFound)
}
```

---

## 3. Implementação no Frontend

### 3.1 Funções Globais no JS (e.g., `app.js`)
Crie sua função de wrapper que evoca a biblioteca:

```javascript
/**
 * Show a toast notification.
 * @param {string} heading - Toast heading (Success / Error).
 * @param {string} text - Toast body text.
 * @param {string} icon - Icon type: 'success' | 'error' | 'warning' | 'info'.
 */
function showToast(heading, text, icon) {
  $.toast({
    heading: heading,
    text: text,
    showHideTransition: 'slide',
    icon: icon || 'info', // O icone determina a cor gerada pela biblioteca.
    position: 'top-right',
    hideAfter: 4000, 
    allowToastClose: true,
  });
}
```

### 3.2 Montando o Base HTML (`base.html` / `layout.html`)

Importe o CSS as libs JS na base do HTML e capture os campos `Flash` e `FlashType` vindos do Render / Parse do Golang. O Golang verifica se `{{.Flash}}` existe e roda o script se for o caso!

```html
<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <title>Nome do App</title>

  <!-- 1. Importando o CSS do Jquery Toast -->
  <link rel="stylesheet" href="/static/css/jquery.toast.min.css">
</head>
<body>
  
  <main>
     {{block "content" .}}{{end}}
  </main>

  <!-- 2. Importe o jQuery primeiro -->
  <script src="/static/js/jquery-4.0.0.min.js"></script>

  <!-- 3. jQuery Toast Plugin -->
  <script src="/static/js/jquery.toast.min.js"></script>
  
  <!-- 4. Seu arquivo JS contendo a função showToast -->
  <script src="/static/js/app.js"></script>

  <!-- 5. Gatilho via Go Templates (Irá disparar assim que a page terminar de carregar) -->
  {{if .Flash}}
  <script>
    $(function() {
      // Usa os dados populados pelo "BaseData" do helper Golang
      // FlashType = Header ("Success"), Flash = Text ("Criado com Sucesso"), FlashIcon = Tema ("success")
      showToast('{{.FlashType}}', '{{.Flash}}', '{{.FlashIcon}}');
    });
  </script>
  {{end}}

</body>
</html>
```

---

## 💡 Fluxo Resumo ("Como a mágica acontece?")
1. O usuário submete um formulário `POST`.
2. O Golang processa o formulário, roda `setFlash(r, w, "success", "Tudo certo!")` (Isto é armazenado de fato nos **cookies** criptografados de Sessão do Header).
3. O Golang responde com um `Redirect 302` pedindo pro Browser voltar pra Página de Tabela.
4. O Browser avança para a página de tabela com GET.
5. O Handler que renderiza a Tabela chama `BaseData()` que por usa vez roda `getFlash()`. Ele lê a mensagem dos cookies e **já exclui a mensagem dali mesmo**.
6. As variavéis são alimentadas para o Go Template Execution (`{{.Flash}} = "Tudo certo!"`).
7. O Go converte o template HTML para output preenchendo o bloco do `<script> showToast...`. 
8. O Browser lê a tag script e exibe o lindo popup verde animado na tela. A próxima atualização de página não acionará novamente, pois a flash session já foi limpa no passo 5!
