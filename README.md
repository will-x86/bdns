# Bad DNS 


The idea here is a social based DNS, where friends share credits / queries.

The goal is not to be a full blown DNS, but rather a beautiful proxy to learn some technologies.


Components: ( tood most LMFAO ) 

- bdns-app Mobile/Desktop app/webapp for managing things etc.
- dns - Core DNS using pingora, goal is to do the following:
    - Parse query ( extract user_id via via sni in DoT/ or subdomain in DoH )
    - Do a simple check if user has the ability to check said side ( credits / full block / whatever)
    - Provide smoochy api for the app



# DNS 
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
