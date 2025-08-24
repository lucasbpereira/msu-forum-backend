# MSU Forum API

Uma API completa para um fórum de perguntas e respostas desenvolvida em Go com Fiber e PostgreSQL.

## 🚀 Funcionalidades

- **Autenticação**: Registro e login com JWT
- **Perguntas**: CRUD completo com tags e votos
- **Respostas**: Sistema de respostas com aceitação
- **Votos**: Sistema de upvote/downvote
- **Tags**: Categorização de perguntas
- **Usuários**: Perfis e gerenciamento
- **Admin**: Painel administrativo

## 📋 Pré-requisitos

- Go 1.24+
- PostgreSQL 12+
- Git

## 🛠️ Instalação

1. **Clone o repositório**
```bash
git clone <url-do-repositorio>
cd msu-forum
```

2. **Instale as dependências**
```bash
go mod tidy
```

3. **Configure o banco de dados**
```bash
# Crie um banco PostgreSQL
createdb msu_forum

# Execute o schema
psql -d msu_forum -f database/schema.sql
```

4. **Configure as variáveis de ambiente**
```bash
# Copie o arquivo de exemplo
cp env.example .env

# Edite o arquivo .env com suas configurações
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=sua_senha_aqui
DB_NAME=msu_forum
APP_PORT=3000
JWT_SECRET=sua_chave_secreta_jwt_aqui_muito_segura
```

5. **Execute a aplicação**
```bash
go run main.go
```

A API estará disponível em `http://localhost:3000`

## 📚 Endpoints da API

### Autenticação
- `POST /register` - Registrar novo usuário
- `POST /login` - Fazer login

### Perguntas (Públicas)
- `GET /questions` - Listar perguntas
- `GET /questions/:id` - Buscar pergunta por ID
- `GET /tags` - Listar tags
- `GET /tags/:id` - Buscar tag por ID
- `GET /tags/:tagId/questions` - Perguntas por tag

### Perguntas (Protegidas)
- `POST /api/questions` - Criar pergunta
- `PUT /api/questions/:id` - Atualizar pergunta
- `DELETE /api/questions/:id` - Deletar pergunta

### Respostas
- `POST /api/questions/:questionId/answers` - Criar resposta
- `GET /api/questions/:questionId/answers` - Listar respostas
- `PUT /api/answers/:id` - Atualizar resposta
- `DELETE /api/answers/:id` - Deletar resposta
- `POST /api/answers/:id/accept` - Aceitar resposta

### Votos
- `POST /api/votes` - Votar em pergunta/resposta
- `GET /api/votes` - Votos do usuário

### Usuários
- `GET /api/profile` - Perfil do usuário
- `PUT /api/profile` - Atualizar perfil
- `GET /api/users/:userId/questions` - Perguntas do usuário
- `GET /api/users/:userId/answers` - Respostas do usuário

### Admin
- `GET /api/admin/users` - Listar usuários
- `PUT /api/admin/users/:userId/status` - Atualizar status do usuário
- `POST /api/admin/tags` - Criar tag
- `PUT /api/admin/tags/:id` - Atualizar tag
- `DELETE /api/admin/tags/:id` - Deletar tag

## 🔐 Autenticação

Para acessar endpoints protegidos, inclua o header:
```
Authorization: Bearer <seu_token_jwt>
```

## 📝 Exemplos de Uso

### Registrar usuário
```bash
curl -X POST http://localhost:3000/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "usuario_teste",
    "email": "teste@example.com",
    "password": "senha123",
    "role": "Member",
    "phone": "11999999999",
    "wallet": "0x123456789"
  }'
```

### Fazer login
```bash
curl -X POST http://localhost:3000/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "teste@example.com",
    "password": "senha123"
  }'
```

### Criar pergunta (com token)
```bash
curl -X POST http://localhost:3000/api/questions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <seu_token>" \
  -d '{
    "title": "Como usar Go com PostgreSQL?",
    "body": "Estou tentando conectar Go com PostgreSQL, alguém pode ajudar?",
    "tags": ["go", "database"]
  }'
```

## 🏗️ Estrutura do Projeto

```
msu-forum/
├── database/
│   ├── database.go      # Conexão com banco
│   └── schema.sql       # Schema do banco
├── handlers/
│   ├── auth_handler.go   # Autenticação
│   ├── question_handler.go # Perguntas
│   ├── answer_handler.go   # Respostas
│   ├── vote_handler.go     # Votos
│   ├── tag_handler.go      # Tags
│   └── user_handler.go     # Usuários
├── middleware/
│   ├── auth.go         # Middleware de autenticação
│   └── cors.go         # Middleware CORS
├── models/
│   ├── user.go         # Modelo de usuário
│   ├── question.go     # Modelo de pergunta
│   ├── answer.go       # Modelo de resposta
│   ├── tag.go          # Modelo de tag
│   └── vote.go         # Modelo de voto
├── main.go             # Arquivo principal
├── go.mod              # Dependências Go
└── README.md           # Este arquivo
```

## 🔧 Configuração de Desenvolvimento

### Variáveis de Ambiente
- `DB_HOST`: Host do PostgreSQL
- `DB_PORT`: Porta do PostgreSQL
- `DB_USER`: Usuário do banco
- `DB_PASSWORD`: Senha do banco
- `DB_NAME`: Nome do banco
- `APP_PORT`: Porta da aplicação
- `JWT_SECRET`: Chave secreta para JWT

### Logs
A aplicação exibe logs no console com informações sobre:
- Conexão com banco de dados
- Inicialização do servidor
- Erros de requisição

## 🚀 Deploy

### Docker (Recomendado)
```bash
# Build da imagem
docker build -t msu-forum .

# Executar container
docker run -p 3000:3000 --env-file .env msu-forum
```

### Produção
1. Configure um servidor PostgreSQL
2. Configure as variáveis de ambiente
3. Execute `go build -o msu-forum main.go`
4. Execute `./msu-forum`

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## 🆘 Suporte

Se você encontrar algum problema ou tiver dúvidas, abra uma issue no repositório.
