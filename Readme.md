## How to use

Start server:

```
make run
```

It will run whole service in docker. It also will setup postgresql + will run migrations.
The server will be started at 8080 port.

## Testing module

1. Start service with `make run`
2. Run test module with `go run ./cmd/usage/main.go`

## Deleting dangle uploads

It's also possible to clean uploads that are not finished uploading for more than 24h. Here how it can be called locally:

```bash
DB_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" go run ./cmd/cleandungle/main.go
```

It should be put as CRON task.

## Useful curls

Send file:
```bash
curl --location 'http://localhost:8080/uploads' \
--form 'file_size="211"' \
--form 'file=@"./test-file.txt"'
```

Get file:
```bash
curl --location 'http://localhost:8080/uploads/{id}'
```

