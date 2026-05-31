# Spec — migration-quasar-tailwind

**Projeto**: go-postfixadmin (frontend Vue 3)  
**Data**: 2026-06  
**Status**: Em andamento  
**Objetivo**: Eliminar a complexidade do Quasar e adotar uma stack mais simples, leve e alinhada ao design neo-brutalist já existente.

---

## Problema

O frontend (Vue 3 + Vite) usava **Quasar Framework v2** para:

- Layout (`q-layout`, `q-header`, `q-drawer`, `q-page-container`)
- Ícones (`q-icon` com Material Icons do `@quasar/extras`)
- Feedback mínimo (`q-spinner`, `q-skeleton`, `q-input`/`q-form` só no login)

A identidade visual "neo-brutalist" (bordas `2-3px solid #1E293B`, `box-shadow: 4px 4px 0`, cantos retos, Fira Sans/Code) **conflitava constantemente** com os defaults Material Design do Quasar, gerando:

- Centenas de seletores `:deep(.q-*)` e overrides globais em `style.css`
- ~1.1MB de CSS do Quasar + fontes de ícones importados
- Dificuldade para evoluir o design

## Solução Adotada

- Remoção completa do Quasar (código + dependências + CSS)
- **Tailwind CSS v4** via `@tailwindcss/vite` (CSS-first, sem `tailwind.config.js`)
- **`@lucide/vue`** — tree-shakeable, já era padrão na versão server-rendered anterior
- Classes base reutilizáveis centralizadas em `style.css` (`.btn-primary`, `.modal-card`, `.spinner`, `.skeleton`, `.error-banner`, `.page-header`)

## Análise do Quasar (estado anterior)

| Categoria       | Componentes                                                    | Frequência        |
|-----------------|----------------------------------------------------------------|-------------------|
| **Layout**      | `q-layout`, `q-header`, `q-drawer`, `q-page-container`, `q-page`, `q-list`, `q-item` | Alta (MainLayout) |
| **Ícones**      | `q-icon` (Material Icons)                                      | ~160+ usos         |
| **Feedback**    | `q-spinner`, `q-skeleton`                                      | Moderada           |
| **Formulários** | `q-form`, `q-input`                                            | Baixa (só Login)   |
| **Outros**      | `q-dialog`, `q-table`, `q-btn`, `q-card`, `q-select` — **não usados** | —            |

## Mapeamento Material → Lucide

| Material Icon                   | Lucide                          |
|---------------------------------|---------------------------------|
| `mail`                          | `Mail`                          |
| `public`                        | `Globe`                         |
| `add_circle`                    | `PlusCircle`                    |
| `edit`                          | `Pencil`                        |
| `delete`                        | `Trash2`                        |
| `warning`                       | `AlertTriangle`                 |
| `shield`                        | `Shield`                        |
| `exit_to_app`                   | `LogOut`                        |
| `visibility` / `visibility_off` | `Eye` / `EyeOff`                |
| `person`                        | `User`                          |
| `lock`                          | `Lock`                          |
| `settings`                      | `Settings`                      |
| `expand_more` / `expand_less`   | `ChevronDown` / `ChevronUp`     |
| `close`                         | `X`                             |
| `save`                          | `Save`                          |
| `refresh`                       | `RefreshCw`                     |
| `arrow_forward`                 | `ArrowRight`                    |
| `history`                       | `History`                       |
| `vpn_key`                       | `Key`                           |
| `content_copy`                  | `Copy`                          |
| `check_circle`                  | `CheckCircle`                   |
| `auto_fix_high`                 | `Wand2` / `Sparkles`            |
| `filter_alt`                    | `Filter`                        |
| `manage_accounts`               | `Users`                         |
| `person_add`                    | `UserPlus`                      |

## Riscos e Mitigações

| Risco                                           | Prob.  | Impacto | Mitigação                                                    |
|-------------------------------------------------|--------|---------|--------------------------------------------------------------|
| `<q-icon>` esquecido em página pouco usada      | Média  | Baixo   | `grep -r 'q-icon' src/` no final da Fase 3                   |
| Diferença sutil de ícones (Material vs Lucide)  | Alta   | Baixo   | Aceitar pequenas diferenças ou ajustar size/stroke           |
| Sidebar perde comportamento "show-if-above"     | Baixa  | Médio   | Sidebar sempre visível no desktop (implementado assim)       |
| Quebra no build de produção                     | Baixa  | Alto    | Build executado após cada fase                               |
| Perda de estilo durante limpeza de CSS          | Média  | Médio   | Limpeza de CSS feita **depois** de todas as páginas migradas |

## Referências

- Gerar lista de ícones: `grep -oP 'name="[^"]+"' src/**/*.vue | sort | uniq`
- Screenshots de referência: `DOCUMENTS/screenshots/`
- Design tokens: `:root` + `@theme` em `frontend/src/style.css`
