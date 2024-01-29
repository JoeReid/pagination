[![Go](https://github.com/JoeReid/pagination/actions/workflows/go.yml/badge.svg)](https://github.com/JoeReid/pagination/actions/workflows/go.yml)

# Pagination

API pagination is a ubiquitous feature and ocasionaly contraversial topic.

This repository provides an opinionated [Specification](SPECIFICATION.md) for performing pagination
along with a golang library for ease of implementation.

The [Specification](SPECIFICATION.md) also adresses the topic of backwards compatibility with other pagination
methods. The golang library also provides helpers for this.

## Usage

```go
middleware := pagination.NewMiddleware()

r := chi.NewRouter()

r.With(middleware).Get("/items", func(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("request: maxItems=%d page=%q\n", pagination.MaxItems(r), pagination.Page(r))

    pagination.SetNext(r, "test") // dummy example token, replace with your own
    w.Write([]byte("...Data..."))
})

http.ListenAndServe(":8080", r)
```

## Pagination flow

```mermaid
sequenceDiagram
    participant Client
    participant Server

    Note left of Client: Request first page
    Client->>+Server: GET https://example.com/items?maxItems=5
    Server-->>Client: Link: <https://example.com/items?maxItems=5&page=3hj53>#59; rel=#34;next#34;
    Server-->>Client: <body>
    deactivate Server

    Note left of Client: Follow link to next page
    Client->>+Server: GET https://example.com/items?maxItems=5&page=3hj53
    Server-->>Client: Link: <https://example.com/items?maxItems=5>#59; rel=#34;prev#34;
    Server-->>Client: <body>
    deactivate Server

    Note left of Client: No next link,<br>pagination finished
```

See the [Specification](SPECIFICATION.md) for a full explanation.
