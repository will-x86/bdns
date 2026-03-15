# Core DNS



See ../README.md for what's implemented




## Commands to remember

### Dev
```bash
docker compose up -d 
air
```

### DB Migrations
#### Creating a migration 
```bash 
GOOSE_MIGRATION_DIR=./migrations goose sqlite3 ./app.db create init sql
```


#### Migrate up 
```bash
GOOSE_MIGRATION_DIR=./migrations goose sqlite3 ./app.db up
```



#### Seed DB
```bash
sqlite3 app.db < seed.sql
```


### Logging
** adding "component" context should not be done by caller, but by .. callee?"** 
e.g.:

```go
// naw / no 
ctx := zerolog.Ctx(ctx).With().Str("component","engine").Logger().WithContext(ctx)
function(ctx)

// yes 
ctx = log.WithContext(ctx)
func function(ctx context.Context){
    log := zerolog.Ctx(ctx).With().Str("component", "engine").Logger()
}

```


env var "production" or "" leads to nice zerolog output
env var "local" leads to pretty printing logs using [Horizontal](https://github.com/UnnoTed/horizontal)

- panic : parser issue - code issue (me ) 
- fatal : want program to panic, primariy for non-stack trace needed errors
- Error: code issue
- Warn: user issues - tls breaking etc
- info: stuff happening .. ?
- Debug: use for helping me
- trace: debugging x2

panic (zerolog.PanicLevel, 5)
fatal (zerolog.FatalLevel, 4)
error (zerolog.ErrorLevel, 3)
warn (zerolog.WarnLevel, 2)
info (zerolog.InfoLevel, 1)
debug (zerolog.DebugLevel, 0)
trace (zerolog.TraceLevel, -1)
