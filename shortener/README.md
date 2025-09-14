<h1 align="center">:zap: URL Shortener - Encurtador de Link :zap:</h1>

<p align="center">
<b>:us:</b> Application used to shorten internet links..<br>
<b>:brazil:</b> Aplicação utilizada para encurtar links da internet.
</p>

<p align="center">
 <a href="#history">História</a> •
 <a href="#objective">Objetivo</a> •
 <a href="#technologies">Tecnologias</a> •
 <a href="#how-to-run">Como Executar</a> •
</p>

---

<h1 id="history">:book: História</h1>

**EN** — A study project in Go focused on building a simple REST API for link shortening. All data is kept in memory (perfect for learning without setting up an external database). You send a URL, the API gives you back a short code, and when that code is visited in the browser, it seamlessly redirects to the original link. This project is great for practicing HTTP basics, routing, handlers, and in-memory storage.

**BR** — Projeto de estudo em Go focado em criar uma API REST simples de encurtamento de links. Os dados ficam em memória (ideal para aprender sem precisar configurar banco externo). Você envia uma URL, recebe um código curto, e ao visitar esse código no navegador é redirecionado para a URL original. Ótimo para revisar HTTP, roteamento, handlers e armazenamento em memória.

---

<h1 id="objective">:bulb: Objetivo | Objective</h1>

### Português
- **Receber uma URL** via requisição POST utilizando o pacote `http`.
- **Gerar um código curto** utilizando o pacote `rand`.
- **Armazenar os dados em memória** usando um simples `map[string]string`.
- **Retornar a URL encurtada** também através do pacote `http`.
- **Quando a URL encurtada for acessada**, redirecionar o usuário para o endereço original.

### English
- **Receive a URL** via POST request using the `http` package.
- **Generate a short code** with the `rand` package.
- **Store the data in memory** using a simple `map[string]string`.
- **Return the shortened URL** through the `http` package.
- **When the shortened URL is accessed**, redirect the user to the original address.

---

<h1 id="technologies">:rocket: Tecnologias</h1>

- **Go** (>= 1.24)
- **chi** (router minimalista para Go)

---

<h1 id="how-to-run">:link: Como Usar a API | How to Use the API</h1>

### Português

#### 1) Encurtar uma URL
- **Endpoint:** `POST /api/shorten`
- **Corpo da requisição (JSON):**

```json
{
  "url": "https://www.google.com"
}
```
Resposta de sucesso (201 Created):
```json
{
  "data": "Ab3k9XyP"
}
```

Neste caso, a URL encurtada ficará disponível em:

http://localhost:8080/Ab3k9XyP


Resposta de erro (exemplo de URL inválida):
```
{
  "error": "invalid url passed"
}
```

2) Acessar uma URL encurtada

Endpoint: GET /{code}

Exemplo:

GET http://localhost:8080/Ab3k9XyP


Comportamento esperado:

O usuário será redirecionado para a URL original (https://www.google.com).

Erro caso o código não exista:
404 Not Found
```json
url nao encontrada
```

Utilizei neste projeto apoio de IA com ChatGPT e Gemini para entender melhor os fluxos da linguagem GO para facilitar meu aprendizado.

<p align="center"> <sub>@jorgediasdsg — 2025</sub> </p>
