# Spec — api-rest

Base path: `/api/v1`  
Auth: `Authorization: Bearer <access_token>` em todos os endpoints protegidos.

---

## Autenticação

| Método | Endpoint            | Auth | Descrição                                              |
|--------|---------------------|------|--------------------------------------------------------|
| POST   | `/auth/login`       | —    | Login (admin ou mailbox). Retorna access token + cookie httpOnly com refresh token |
| POST   | `/auth/refresh`     | —    | Renova access token via cookie httpOnly                |
| POST   | `/auth/logout`      | JWT  | Invalida sessão                                        |
| GET    | `/auth/me`          | JWT  | Dados do usuário autenticado                           |

---

## Recursos (Admin)

### Domains
| Método | Endpoint             | Notas           |
|--------|----------------------|-----------------|
| GET    | `/domains`           | Lista (scoped)  |
| POST   | `/domains`           | Criar           |
| GET    | `/domains/:domain`   | Detalhe         |
| PUT    | `/domains/:domain`   | Editar          |
| DELETE | `/domains/:domain`   | Remover         |

### Mailboxes
| Método | Endpoint                  | Notas           |
|--------|---------------------------|-----------------|
| GET    | `/mailboxes`              | Lista (scoped)  |
| POST   | `/mailboxes`              | Criar           |
| GET    | `/mailboxes/:username`    | Detalhe         |
| PUT    | `/mailboxes/:username`    | Editar          |
| DELETE | `/mailboxes/:username`    | Remover         |

### Aliases
| Método | Endpoint               | Notas           |
|--------|------------------------|-----------------|
| GET    | `/aliases`             | Lista (scoped)  |
| POST   | `/aliases`             | Criar           |
| GET    | `/aliases/:address`    | Detalhe         |
| PUT    | `/aliases/:address`    | Editar          |
| DELETE | `/aliases/:address`    | Remover         |

### Alias Domains
| Método | Endpoint                          | Notas           |
|--------|-----------------------------------|-----------------|
| GET    | `/alias-domains`                  | Lista (scoped)  |
| POST   | `/alias-domains`                  | Criar           |
| GET    | `/alias-domains/:alias_domain`    | Detalhe         |
| PUT    | `/alias-domains/:alias_domain`    | Editar          |
| DELETE | `/alias-domains/:alias_domain`    | Remover         |

### Admins
| Método | Endpoint               | Notas              |
|--------|------------------------|--------------------|
| GET    | `/admins`              | Lista (superadmin) |
| POST   | `/admins`              | Criar              |
| GET    | `/admins/:username`    | Detalhe            |
| PUT    | `/admins/:username`    | Editar             |
| DELETE | `/admins/:username`    | Remover            |

### Transports
| Método | Endpoint              | Notas              |
|--------|-----------------------|--------------------|
| GET    | `/transports`         | Lista (superadmin) |
| POST   | `/transports`         | Criar              |
| GET    | `/transports/:id`     | Detalhe            |
| PUT    | `/transports/:id`     | Editar             |
| DELETE | `/transports/:id`     | Remover            |

### API Keys
| Método | Endpoint                    | Notas   |
|--------|-----------------------------|---------|
| GET    | `/settings/apikeys`         | Lista   |
| POST   | `/settings/apikeys`         | Criar   |
| PUT    | `/settings/apikeys/:id`     | Editar  |
| DELETE | `/settings/apikeys/:id`     | Remover |

---

## Recursos (User self-service)

| Método | Endpoint              | Descrição                   |
|--------|-----------------------|-----------------------------|
| GET    | `/user/me`            | Perfil do usuário           |
| GET    | `/user/forwarding`    | Configuração de forwarding  |
| POST   | `/user/forwarding`    | Atualizar forwarding        |
| POST   | `/user/password`      | Trocar senha                |
| GET    | `/user/vacation`      | Configuração de férias      |
| POST   | `/user/vacation`      | Ativar/atualizar férias     |
| DELETE | `/user/vacation`      | Desativar férias            |

---

## Relatórios e Dashboard

| Método | Endpoint       | Descrição                         |
|--------|----------------|-----------------------------------|
| GET    | `/dashboard`   | Estatísticas gerais               |
| GET    | `/logs`        | Logs de admin (paginado, scoped)  |
| GET    | `/maillog`     | Mail log (paginado, scoped)       |
| GET    | `/version`     | Versão da aplicação               |

---

## OpenAPI / Swagger

Spec disponível em: `GET /swagger/doc.json`  
Arquivo estático: [docs/swagger.json](../../../docs/swagger.json)

Visualizar com:
- `http://localhost:8080/swagger/index.html` (Swagger UI embutida)
- [editor.swagger.io](https://editor.swagger.io/)

---

## Como rodar

```bash
./postfixadmin server
# ou
go run main.go server
```
