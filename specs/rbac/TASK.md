# Tasks — rbac

## Em andamento

_(nenhuma tarefa ativa no momento)_

## Pendente

### Frontend — Gerenciamento de Roles (Vue 3)

- [ ] Criar página `frontend/src/views/RoleManagement.vue` — lista roles, cria roles customizados, edita permissions
- [ ] Criar componente `frontend/src/components/RoleAssignment.vue` — painel de atribuição de roles integrado ao modal de edição de admin
- [ ] Atualizar `frontend/src/stores/auth.ts` com método `hasPermission(perm: string): boolean` lendo `Claims.Permissions` do JWT decodificado
- [ ] Adicionar guards de rota em `frontend/src/router/index.ts` — redirecionar para 403 quando `hasPermission` falhar
- [ ] Adicionar item "Roles" no menu lateral (`AppSidebar.vue` ou equivalente) — visível apenas para superadmin

### Backend — Itens restantes

- [ ] Subcomando CLI `rbac assign <username> <role> [domain]` — atribuir role via terminal sem precisar da UI
- [ ] Testes unitários para `RequirePermission` — cobrir: superadmin bypass, wildcard, permission match, domain scope, feature flag off

### Operacional

- [ ] Documentar query SQL para atribuir `domain_admin` a admins existentes na wiki/README
- [ ] Criar script/seed para atribuir role automaticamente a admins que já têm `domain_admins` (evitar inserção manual)
