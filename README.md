# Pranav's Nginx Log Parser

## Purpose
Nginx log files are important as they can help diagnose issues, look for potential threat actors, and provide telemetry about your NGINX server. Scrolling through logs manually is quite inconvienient especially when dealing with hundreds of archived logs. To address this, I built a simple Go CLI to help make parsing these log files easier.

This is a project I did to learn Go. There's probably something better than this that exists so uh don't take this too seriously. I built a log parser because I want to work on infrastructure security and observability.

Test files sourced from my personal VPS's Nginx logs.

## Download/Run
First, ensure you have Go installed
```bash
git clone https://github.com/s9kt/go-log-parser
cd go-log-parser
go run .
```

## Build

Follow the same steps as above but do `go build .` to build a binary instead.

## Usage
`log-parser [FLAGS]`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-file` | string | stdin | Path to the log file |
| `-format` | string | `nginx` | Log format (`nginx`, `nginx-error`, `json`) |
| `-filter` | string | | Filter by field using `field=value` syntax |
| `-aggregate` | string | | Field to aggregate and count |
| `-top` | int | `10` | Limit aggregate output. Use `-1` for all results |
| `-verbose` | bool | `false` | Print unparseable lines to stderr |

## Examples
```
cat ./testdata/access.log | go run . --aggregate status

status amount
404 609
200 148
301 67
405 16
400 11
302 5
308 2
499 1
304 1
101 1
```
```
go run . --file ./testdata/error.log --format nginx-error --filter client=45.148.10.5

2026/05/16 19:55:03 [error] 1123366#1123366: *15131 open() "/usr/share/nginx/html/asdkjh2k3h4_nonexistent_path" failed (2: No such file or directory), client: 45.148.10.5, server: localhost, request: "GET /asdkjh2k3h4_nonexistent_path HTTP/1.1", host: "199.195.249.151"
2026/05/16 19:55:03 [error] 1123366#1123366: *15131 open() "/usr/share/nginx/html/.env" failed (2: No such file or directory), client: 45.148.10.5, server: localhost, request: "GET /.env HTTP/1.1", host: "199.195.249.151"
2026/05/16 19:55:03 [error] 1123366#1123366: *15131 open() "/usr/share/nginx/html/.git/config" failed (2: No such file or directory), client: 45.148.10.5, server: localhost, request: "GET /.git/config HTTP/1.1", host: "199.195.249.151"
2026/05/16 19:55:03 [error] 1123366#1123366: *15131 open() "/usr/share/nginx/html/.env.local" failed (2: No such file or directory), client: 45.148.10.5, server: localhost, request: "GET /.env.local HTTP/1.1", host: "199.195.249.151"
```
```
go run . --file ./testdata/error.log --format nginx-error --filter time_local="2026/05/16 19:55:19"
2026/05/16 19:55:19 [error] 1123366#1123366: *15131 open() "/usr/share/nginx/html/.pypirc" failed (2: No such file or directory), client: 45.148.10.5, server: localhost, request: "GET /.pypirc HTTP/1.1", host: "199.195.249.151"
2026/05/16 19:55:19 [error] 1123366#1123366: *15131 open() "/usr/share/nginx/html/.vscode/settings.json" failed (2: No such file or directory), client: 45.148.10.5, server: localhost, request: "GET /.vscode/settings.json HTTP/1.1", host: "199.195.249.151"
2026/05/16 19:55:19 [error] 1123366#1123366: *15131 open() "/usr/share/nginx/html/.idea/workspace.xml" failed (2: No such file or directory), client: 45.148.10.5, server: localhost, request: "GET /.idea/workspace.xml HTTP/1.1", host: "199.195.249.151"
2026/05/16 19:55:19 [error] 1123366#1123366: *15131 open() "/usr/share/nginx/html/.github/workflows/main.yml" failed (2: No such file or directory), client: 45.148.10.5, server: localhost, request: "GET /.github/workflows/main.yml HTTP/1.1", host: "199.195.249.151"
```
```
go run . --file ./testdata/access.log --aggregate time_local --top 7
time_local amount

16/May/2026:08:49:26 -0400 27
16/May/2026:08:49:29 -0400 25
16/May/2026:12:45:03 -0400 21
16/May/2026:18:30:48 -0400 19
16/May/2026:16:45:25 -0400 13
16/May/2026:13:52:33 -0400 13
16/May/2026:13:52:31 -0400 13
```
