# Spec: admin-panel-ui

Layout ZimbraAdmin Classic em Vue 3, servido pelo binário único em `:7071`. Referência: `https://192.168.56.30:7071/zimbraAdmin/`. Auditor de paridade: agente `qa-frontend-cloner` (reusado do go-snappymail).

## ADDED Requirements

### Requirement: Interface toda em inglês
Todo o texto da UI do painel admin (login, árvore de navegação, telas, toolbars, labels, mensagens de erro, toasts) SHALL estar em **inglês** (ex.: "Home", "Manage", "Accounts", "Domains", "New", "Delete", "Username", "Password", "Login"). A referência legada aparece em português (locale da VM), mas o clone usa inglês.

#### Scenario: Strings em inglês
- **WHEN** qualquer tela do painel renderiza
- **THEN** todos os rótulos e mensagens estão em inglês, sem strings localizadas em outro idioma

### Requirement: Painel admin servido em :7071
O binário SHALL servir a SPA de admin em `:7071` (http, ou https quando `[admin] tls=true`), habilitado por `[admin] enabled=true`, sem colidir com o webmail (`:8082`).

#### Scenario: Admin sobe isolado
- **WHEN** o subcomando `server` roda com `[admin] enabled=true`
- **THEN** o painel responde em `:7071` e o webmail (se habilitado) em `:8082`, cada um com seu roteador/SPA e cookies isolados (nome/path/SameSite próprios)

### Requirement: Tela de login idêntica ao ZimbraAdmin
A tela de login SHALL clonar o login do ZimbraAdmin Classic (referência `192.168.56.30:7071`, print em `docs/prints/zimbra/41-zimbraadmin-login.png`): card centrado com topo em **gradiente azul** e marca textual "Admin Console" (sem logo Zimbra), campos Nome do usuário / Senha com **labels à esquerda** e inputs brancos, botão **Login** pequeno no canto inferior direito, reflexo/rodapé cinza abaixo do card, sobre página cinza-claro; linha de copyright no rodapé. Distinta do login do webmail.

#### Scenario: Render do login admin
- **WHEN** um usuário não autenticado abre `:7071`
- **THEN** vê o card com topo azul "Admin Console", labels à esquerda, botão Login no canto inferior direito e o rodapé — visualmente equivalente ao console real (marca textual, motor Vue como divergências intencionais)

#### Scenario: Autenticação
- **WHEN** o admin envia usuário/senha válidos com papel admin
- **THEN** entra no painel; credenciais inválidas ou sem papel admin mostram erro e não autenticam

### Requirement: Top bar ZimbraAdmin
A barra superior SHALL clonar o console: faixa azul com marca textual "Administration", busca central, menu `admin@… ▾`, botão refresh. Sem logo Zimbra.

#### Scenario: Render da top bar
- **WHEN** o admin autenticado abre o painel
- **THEN** a barra azul, a busca e o menu de usuário aparecem na mesma disposição do console real

### Requirement: Árvore de navegação
A navegação lateral SHALL ter a estrutura do ZimbraAdmin: Home / Monitor / Manage (Accounts, Aliases, Distribution Lists, Resources, Domains) / Configure (Class of Service, Servers, Global Settings, Zimlets) / Tools & Migration / Search. Os nós **com backend** (Accounts, Domains, Aliases, Admins) são funcionais; os **sem fonte** no schema/RBAC atual (Resources, COS, Servers, Global Settings, Zimlets, Monitor) renderizam como **stub "coming soon"/desabilitado**, nunca com dados inventados. Visibilidade por permissão RBAC (ver `admin-backend-api`).

#### Scenario: Navegar entre seções
- **WHEN** o admin clica num nó funcional da árvore
- **THEN** o content pane carrega a tela e o nó fica selecionado

#### Scenario: Nó sem backend
- **WHEN** o admin abre um nó fora de escopo (ex.: Zimlets, COS)
- **THEN** vê um placeholder "coming soon", sem dados fabricados

### Requirement: Home overview
A Home SHALL exibir "Overview" com os contadores reais (Accounts, Domains, Aliases, Admins) e cards de setup. Campos do console legado sem fonte (Version do produto, Servers, COS, Service status, Active sessions, Queue) aparecem como `—`/"n/a" ou ocultos — nunca fabricados.

#### Scenario: Contagens reais
- **WHEN** a Home carrega
- **THEN** os contadores batem com `GET /api/v1/admin/overview` (accounts, domains); campos sem fonte mostram "n/a"

### Requirement: List view + toolbar
As telas de gerenciamento SHALL usar list view estilo Zimbra (colunas, seleção com outline, paginação) e toolbar (New/Edit/Delete conforme a tela).

#### Scenario: Selecionar e agir
- **WHEN** o admin seleciona uma linha
- **THEN** as ações da toolbar habilitam e operam via API do backend

### Requirement: Paridade visual validada
O layout SHALL ser iterado com o `qa-frontend-cloner` contra `:7071` até paridade (paleta, tipografia, cantos 3px, densidade), permitindo divergências intencionais (marca textual, motor Vue).

#### Scenario: Auditoria de paridade
- **WHEN** o QA compara telas do painel com o console real
- **THEN** a auditoria não retorna P1 (divergências intencionais listadas)
