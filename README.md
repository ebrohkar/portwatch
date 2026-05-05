# portwatch

Lightweight daemon that monitors open ports and alerts on unexpected changes with configurable rules.

## Installation

```bash
go install github.com/yourusername/portwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/portwatch.git && cd portwatch && go build -o portwatch .
```

## Usage

Start the daemon with a config file:

```bash
portwatch --config /etc/portwatch/config.yaml
```

Example `config.yaml`:

```yaml
interval: 30s
alert:
  method: log
  path: /var/log/portwatch.log
rules:
  - port: 22
    allowed: true
  - port: 80
    allowed: true
  - port: 443
    allowed: true
  - port: "*"
    allowed: false
    notify: true
```

portwatch will scan open ports at the configured interval and trigger alerts whenever a port outside your ruleset is detected.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `./config.yaml` | Path to configuration file |
| `--interval` | `30s` | Override scan interval |
| `--dry-run` | `false` | Scan once and print results without alerting |

## License

MIT © portwatch contributors