# awx-syslog

Ansible AWX does not support syslog forwarding natively, but it does support a generic HTTP endpoint where it sends JSON logs (https://docs.ansible.com/automation-controller/4.4/html/administration/logging.html#logging-aggregator-services).

This project is a simple HTTP server that listens for AWX JSON logs and forwards them to syslog.

## Installation

Download the latest binary from the [releases page](https://github.com/juanfont/awx-syslog/releases) and place it in your PATH (e.g. `/usr/local/bin/awx-syslog`).

Don't forget to make it executable (`chmod +x /usr/local/bin/awx-syslog`).

Add the following configuration to in `/etc/awx-syslog/config.yaml`:

```yaml
listen_addr: "0.0.0.0:8080"
log_level: "info"
hostname_field: "awx.example.com"
syslog:
  server_addr: "10.0.0.1:514"
  protocol: "udp"
```

Create the following `/etc/systemd/system/awx-syslog.service`:

```ini
[Unit]
Description=AWX Syslog
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/awx-syslog serve # or whatever path you downloaded the binary to
Restart=always
RestartSec=5
Environment='AWX_SYSLOG_CONFIG=/etc/awx-syslog/config.yaml'

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable awx-syslog
sudo systemctl start awx-syslog
```
