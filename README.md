# inifmt

`inifmt` is a command-line tool for formatting INI configuration files. It aligns the `=` signs in key-value pairs, making configuration files more readable.

## Features

- Aligns equals signs (`=`) in key-value pairs.
- Operates on the entire file or on a per-section basis.
- Optionally aligns comments and blank lines.
- Single-space formatting mode ensuring exactly one space around `=`.

## Installation

To install, clone the repository and build/install (adjust according to your environment):

```bash
git clone <repository-url>
cd inifmt
make install
```

## Usage

Run `inifmt` from the command line. Without a filename, it reads from stdin:

```bash
inifmt [file]
```

Use the `-w` or `--write` flag to overwrite the file with formatted content. For additional help, run:

```bash
inifmt -h
```

## Examples

**Align entire file:**

```bash
inifmt input.ini > output.ini
```

**Per-section alignment including comments:**

```bash
inifmt --per-section -C input.ini -w
```

**Single-space formatting mode:**

```bash
inifmt --single-space input.ini > output.ini
```

## Flags

- `-w`, `--write`: Write changes back to the file (when a filename is provided).
- `-s`, `--per-section`: Align `=` signs within each section independently.
- `-u`, `--single-space`: Ensure exactly one space around `=` signs.
- `-C`, `--include-comments`: Include comments in the alignment process.

## License

This project is licensed under the [MIT License](LICENSE).
