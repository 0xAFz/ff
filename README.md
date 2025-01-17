# FF

Fast Packet Forwarder (L4 Reverse Proxy)

## Supported Protocols
- UDP

## Build
```bash
go build -o ff main.go
```

## Usage
```bash
ff --addr 127.0.0.1:8000 --dest 127.0.0.1:8080
```

