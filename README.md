# import-fidelity

Download data from Fidelity's website. Supported downloads include:

1. Ticker information (Stock type, currency, exchange, symbol, name, CUSIP, and CIK)
2. Account activity

# Install

1. compile the software

```bash
mage build
```

2. install playwright OS dependencies

```bash
go install github.com/playwright-community/playwright-go/cmd/playwright@latest
playwright install --with-deps chromium
```

# Configuration

`import-fidelity` uses viper for configuration. This enables a
variety of mechanisms for providing configuration. Currently supported
methods are:

 1. Environment variables
 2. TOML file
 3. Command line flags

For a complete list of configuration parameters run `import-fidelity --help`

## Exit codes

32 - Activity page error
33 - Backblaze error
34 - Login error
35 - Write parquet