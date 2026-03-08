# linear-cli

Command-line interface for the Linear project management API.

## Installation

### From source

Requirements: Go 1.25+

```
git clone https://github.com/your-org/linear-cli
cd linear-cli
make build
```

The binary is placed at `./linear-cli`. Move it to a directory in your PATH:

```
mv linear-cli /usr/local/bin/linear
```

## Authentication

linear-cli requires a Linear personal API key.

Generate one at: Linear Settings -> API -> Personal API keys

### Save key to config

```
linear auth
```

Prompts for your API key and saves it to the config file.

### Check auth status

```
linear auth status
```

Shows whether an API key is configured (masked).

## Configuration

Config file location: `~/.config/linear-cli/config.yaml`

Format:

```yaml
api_key: lin_api_xxxx
default_team: ""
```

### Environment variable override

Set `LINEAR_API_KEY` to override the config file value without modifying it:

```
export LINEAR_API_KEY=lin_api_xxxx
linear auth status
```

The environment variable takes precedence over the config file.

## Usage

```
linear [command] [flags]

Flags:
  --json        Output in JSON format
  --version     Show version
  -h, --help    Show help
```
