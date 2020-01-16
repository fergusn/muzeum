# Muzeum - a modern OSS artifact repository

Muzeum is a artifact repository that support local and remote repositories.

- Support for Docker, Debian and NuGet - more coming soon
- HTTP(S) proxy with TLS interception 
- Package metrics published via Prometheus endpoint

## Getting Started

```bash

# Muzeum can generate a CA certificate when used as an TLS interception proxy
# This certificate need to be added to the trusted CAs on clients
> muzeum ca                             

# See configuration in examples
> muzeum server --config config.yaml 

# Configure Docker to use proxy - Muzeum will cache imaged from configured registries
> mkdir -p /etc/systemd/system/docker.service.d/
> printf "[Service]\nEnvironment=\"HTTPS_PROXY=https://localhost:8443/\"" > /etc/systemd/system/docker.service.d/https-proxy.conf
> systemctl daemon-reload && systemctl restart docker 

# Confige apt to use the proxy - muzeum will cache downloaded packages
> mkdir -p /etc/apt/apt.conf.d/proxy.conf/
> printf "Acquire::http::Proxy \"http://localhost:8080/\";" > /etc/apt/apt.conf.d/proxy.conf
> apt update

```


 