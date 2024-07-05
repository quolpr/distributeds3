# How to use

Start server:

```
make run
```

It will run whole service in docker. It also will setup postgresql + will run migrations.
The server will be started at 8080 port.

## Curls

Send file:
```bash
curl --location 'http://localhost:8080/uploads' \
--form 'file_size="211"' \
--form 'file=@"./test-file.txt"'
```

Get file:
```bash
curl --location 'http://localhost:8080/uploads/5fb47688-4468-4b08-bd51-3d578794be29'
```
