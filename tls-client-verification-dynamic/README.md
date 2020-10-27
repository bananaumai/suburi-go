# Example for the custom client certificate verification for TLS

You can define your own client certificate verification for TLS connection,
if you appropriately configure [`tls.Config`](https://golang.org/pkg/crypto/tls/#Config):
specifically, you need to define `VerifyPeerCertificate` callback appropriately.

## How to run the example apps

server

```
$ go run cmd/server/main.go -serverCert=data/certs/server -clientCert=data/certs/project1
```

client

```
$ go run cmd/client/main.go -serverCA=data/certs/server/ca.crt -clientCert=data/certs/project1
```

## Test with your own certificate

### How to create client certificates

You need to create `ca.crt`(CA), `client.crt`(client cert), and `client.key`(private key corresponding to client cert).

```
$ openssl genrsa -out ca.key 2048 
$ openssl req -x509 -new -nodes -key ca.key -out ca.crt -sha256 -days 1000
$ openssl genrsa -out client.key 2048
$ openssl req -new -key client.key -out client.csr
$ openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -sha256 -days 500 
```

### How to create server certs

You need to create `ca.crt`(CA), `server.crt`(server cert), and `server.key`(private key corresponding to server cert).

```
$ openssl genrsa -out ca.key 2048 
$ openssl req -x509 -new -nodes -key ca.key -out ca.crt -sha256 -days 1000
$ openssl genrsa -out server.key 2048
$ openssl req -new -key server.key -out server.csr
$ cat << END > server.ext
subjectAltName = @alt_names
[ alt_names ]
DNS.1 = localhost
IP.1 = 127.0.0.1
END
$ openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -sha256 -days 500 -extfile server.ext
```

Please note, you need to specify `Subject Alternative Name`(SAN) so that the HTTP clients can get access to the server.
