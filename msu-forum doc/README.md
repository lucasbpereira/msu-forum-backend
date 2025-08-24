# MSU Forum API - Rotas para Bruno

Este diretório contém todas as rotas da API organizadas para uso no Bruno.

## 📁 Estrutura das Pastas

```
msu-forum doc/
├── Auth/                    # Autenticação
│   ├── Register.bru        # Registrar usuário
│   └── Login.bru           # Fazer login
├── Questions/              # Perguntas
│   ├── ListQuestions.bru   # Listar perguntas (pública)
│   ├── GetQuestion.bru     # Buscar pergunta (pública)
│   ├── CreateQuestion.bru  # Criar pergunta (protegida)
│   ├── UpdateQuestion.bru  # Atualizar pergunta (protegida)
│   └── DeleteQuestion.bru  # Deletar pergunta (protegida)
├── Answers/                # Respostas
│   ├── CreateAnswer.bru    # Criar resposta (protegida)
│   ├── ListAnswers.bru     # Listar respostas (protegida)
│   ├── UpdateAnswer.bru    # Atualizar resposta (protegida)
│   ├── DeleteAnswer.bru    # Deletar resposta (protegida)
│   └── AcceptAnswer.bru    # Aceitar resposta (protegida)
├── Votes/                  # Votos
│   ├── Vote.bru            # Votar (protegida)
│   └── GetUserVotes.bru    # Votos do usuário (protegida)
├── Tags/                   # Tags
│   ├── ListTags.bru        # Listar tags (pública)
│   ├── GetTag.bru          # Buscar tag (pública)
│   ├── GetQuestionsByTag.bru # Perguntas por tag (pública)
│   ├── CreateTag.bru       # Criar tag (admin)
│   ├── UpdateTag.bru       # Atualizar tag (admin)
│   └── DeleteTag.bru       # Deletar tag (admin)
├── Users/                  # Usuários
│   ├── GetProfile.bru      # Perfil do usuário (protegida)
│   ├── UpdateProfile.bru   # Atualizar perfil (protegida)
│   ├── GetUserQuestions.bru # Perguntas do usuário (protegida)
│   └── GetUserAnswers.bru  # Respostas do usuário (protegida)
├── Admin/                  # Administração
│   ├── ListUsers.bru       # Listar usuários (admin)
│   └── UpdateUserStatus.bru # Atualizar status (admin)
├── environments/
│   └── local.bru           # Ambiente local
└── README.md               # Este arquivo
```

## 🚀 Como Usar

### 1. Configuração Inicial

1. **Abra o Bruno** e importe esta pasta
2. **Selecione o ambiente "local"** no canto superior direito
3. **Configure a variável `authToken`** (será preenchida após login)

### 2. Fluxo de Teste Recomendado

#### Passo 1: Autenticação
1. Execute `Auth/Register.bru` para criar um usuário
2. Execute `Auth/Login.bru` para fazer login
3. **Copie o token** da resposta do login
4. **Cole o token** na variável `authToken` do ambiente

#### Passo 2: Testar Rotas Públicas
1. `Questions/ListQuestions.bru` - Listar perguntas
2. `Questions/GetQuestion.bru` - Buscar pergunta específica
3. `Tags/ListTags.bru` - Listar tags
4. `Tags/GetTag.bru` - Buscar tag específica

#### Passo 3: Testar Rotas Protegidas
1. `Questions/CreateQuestion.bru` - Criar pergunta
2. `Answers/CreateAnswer.bru` - Criar resposta
3. `Votes/Vote.bru` - Votar em pergunta/resposta
4. `Users/GetProfile.bru` - Ver perfil

#### Passo 4: Testar Rotas Admin (se for admin)
1. `Admin/ListUsers.bru` - Listar usuários
2. `Tags/CreateTag.bru` - Criar tag
3. `Admin/UpdateUserStatus.bru` - Atualizar status de usuário

## 🔐 Tipos de Autenticação

### Rotas Públicas
- Não requerem autenticação
- Podem ser acessadas por qualquer pessoa

### Rotas Protegidas
- Requerem token JWT no header `Authorization: Bearer {{authToken}}`
- Apenas usuários logados podem acessar

### Rotas Admin
- Requerem token JWT
- Apenas usuários com role "Admin" podem acessar

## 📝 Exemplos de Uso

### Criar uma Pergunta
1. Faça login e configure o `authToken`
2. Execute `Questions/CreateQuestion.bru`
3. Modifique o body JSON conforme necessário:
```json
{
  "title": "Minha pergunta",
  "body": "Conteúdo da pergunta",
  "tags": ["go", "api"]
}
```

### Votar em uma Pergunta
1. Execute `Votes/Vote.bru`
2. Modifique o body JSON:
```json
{
  "post_type": "question",
  "post_id": 1,
  "type": 1
}
```

## ⚠️ Observações Importantes

1. **IDs Dinâmicos**: Os IDs nas URLs (como `/questions/1`) devem ser substituídos pelos IDs reais retornados pela API
2. **Token Expiração**: O token JWT expira em 24 horas
3. **Permissões**: Verifique se seu usuário tem as permissões necessárias para rotas admin
4. **Dados de Teste**: Os exemplos usam dados fictícios, ajuste conforme necessário

## 🔧 Troubleshooting

### Erro 401 (Unauthorized)
- Verifique se o `authToken` está configurado
- Faça login novamente se o token expirou

### Erro 403 (Forbidden)
- Verifique se seu usuário tem role "Admin" para rotas admin
- Verifique se você é o autor do conteúdo para edições

### Erro 404 (Not Found)
- Verifique se os IDs nas URLs estão corretos
- Verifique se o recurso existe no banco de dados

## 📊 Status Codes

- `200` - Sucesso
- `201` - Criado com sucesso
- `400` - Dados inválidos
- `401` - Não autorizado
- `403` - Proibido
- `404` - Não encontrado
- `409` - Conflito (ex: tag já existe)
- `500` - Erro interno do servidor
