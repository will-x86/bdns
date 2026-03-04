# Bad DNS 






## Usage

### Prereq
```bash
cargo install cargo-watch
```
### Running (dev)
```bash
RUST_LOG=INFO cargo watch -x "run -- -c conf.yaml"

```

## Rust version


Currently using rustc 1.93.1 though [Pingora](https://github.com/cloudflare/pingora) currently (2nd March 2026) uses 1.84 





## Running in background

```bash
cargo run -- -d
```


### Stopping running in background:

```bash
pkill -SIGTERM bdns
```
