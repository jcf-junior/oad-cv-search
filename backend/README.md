# CV Search — Backend

API em Go para pesquisa semântica de CVs. Usa PostgreSQL com pgvector e o Ollama (modelo `bge-m3`) para gerar embeddings.

## Pré-requisitos

- Go 1.26 ou superior
- Docker Desktop
- Ollama (apenas se `OLLAMA_DOCKER=false`)

## Instalação

1. Clonar o repositório:

   ```
   git clone <URL-DO-REPOSITORIO>
   cd cv-search/backend
   ```

2. Definir o ficheiro `.env`:

   ```
   cp .env.example .env
   ```

   A variável `OLLAMA_DOCKER` define onde o Ollama corre:

   - `true`: o Ollama corre no Docker. Não é necessário instalar nada.
   - `false`: o Ollama corre no sistema local. Siga os passos 3 e 4.

   Se alterar `OLLAMA_DOCKER`, o valor de `COMPOSE_PROFILES` acompanha-o automaticamente.

3. Instalar o Ollama (apenas se `OLLAMA_DOCKER=false`): https://ollama.com/download

4. Instalar o modelo (apenas se `OLLAMA_DOCKER=false`):

   ```
   ollama pull bge-m3
   ```

   Com `OLLAMA_DOCKER=true`, o Docker descarrega o modelo no passo seguinte. A primeira execução demora alguns minutos.

5. Arrancar a infraestrutura:

   ```
   docker compose up -d
   ```

   Este comando arranca o PostgreSQL com pgvector, aplica as migrações e, se `OLLAMA_DOCKER=true`, arranca o Ollama e descarrega o modelo.

6. Instalar as dependências Go:

   ```
   go mod tidy
   ```

7. Arrancar a API:

   ```
   go run .
   ```

   A API fica disponível em `http://localhost:8080`. A mesma porta serve o frontend: abra `http://localhost:8080` no browser. A API tem de arrancar a partir da pasta `backend/`, para que `../frontend` exista.

## Utilização

### Upload de um CV

```
curl -X POST http://localhost:8080/candidates -F "file=@cv.pdf"
```

Resposta (201): candidato com `id`, `name`, `file_name`, `content` e `created_at`.

### Pesquisa semântica

```
curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{"query": "Platform Engineer com experiência em Kubernetes, Terraform e CI/CD", "limit": 5}'
```

Resposta (200): lista de candidatos ordenada por `similarity` (de 0 a 1, maior é mais semelhante). O campo `limit` é opcional (5 por omissão, máximo 20).

### Pesquisa tradicional (palavras-chave)

```
curl -X POST http://localhost:8080/search/keyword \
  -H "Content-Type: application/json" \
  -d '{"query": "kubernetes terraform", "limit": 5}'
```

Usa full-text search do PostgreSQL (configuração `simple`). Todas as palavras da query têm de existir no CV. Resposta (200): lista ordenada por `rank` (`ts_rank`). Serve de comparação com a pesquisa semântica.

### Número de CVs na base

```
curl http://localhost:8080/candidates/count
```

Resposta (200): `{"count": 147}`.

### Detalhe de um candidato

```
curl http://localhost:8080/candidates/<id>
```

### PowerShell

No PowerShell, use `curl.exe` em vez de `curl`. Para a pesquisa:

```
Invoke-RestMethod -Method Post -Uri http://localhost:8080/search -ContentType "application/json" -Body '{"query": "Platform Engineer com Kubernetes", "limit": 5}'
```

## Variáveis de ambiente

| Variável | Utilização | Valor por omissão |
| --- | --- | --- |
| `POSTGRES_USER` | Utilizador do PostgreSQL (Docker) | `cvsearch` |
| `POSTGRES_PASSWORD` | Palavra-passe do PostgreSQL (Docker) | `cvsearch` |
| `POSTGRES_DB` | Nome da base de dados (Docker) | `cvsearch` |
| `OLLAMA_URL` | URL do Ollama usado pela API | `http://localhost:11434` |
| `OLLAMA_MODEL` | Modelo de embeddings | `bge-m3` |
| `OLLAMA_DOCKER` | `true` corre o Ollama no Docker | `true` |
| `COMPOSE_PROFILES` | Perfil do Compose, derivado de `OLLAMA_DOCKER` | `ollama-${OLLAMA_DOCKER}` |
| `DATABASE_URL` | Ligação da API à base de dados | `postgres://cvsearch:cvsearch@localhost:5432/cvsearch?sslmode=disable` |
| `PORT` | Porta da API | `8080` |
| `FRONTEND_DIR` | Pasta do frontend servida pela API | `../frontend` |

A API lê `DATABASE_URL`, `OLLAMA_URL`, `OLLAMA_MODEL` e `PORT` do ambiente do processo. Se não estiverem definidas, usa os valores por omissão. O Docker Compose lê o ficheiro `.env`.

## Estrutura

```
backend/
  main.go
  docker-compose.yml
  migrations/            migrações SQL
  internal/
    config/              variáveis de ambiente
    handler/             HTTP
    service/             extração de PDF, embeddings, lógica
    repository/          SQL e pgvector
    model/               structs
    ollama/              cliente do Ollama
```

## Parar a infraestrutura

```
docker compose down
```

Para apagar também os dados: `docker compose down -v`.
