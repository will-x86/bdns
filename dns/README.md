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
