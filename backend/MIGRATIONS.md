# Database Migrations Strategy

## Current Approach: GORM AutoMigrate

The application currently uses GORM's AutoMigrate feature, which automatically creates and updates database tables based on model definitions.

### How It Works

On server startup ([cmd/server/main.go:45](backend/cmd/server/main.go#L45)):
```go
if err := database.AutoMigrate(); err != nil {
    log.Fatalf("Failed to run migrations: %v", err)
}
```

### Pros
- Simple development workflow
- No migration files to manage
- Automatically creates new columns and indexes

### Cons
- Cannot drop columns or tables automatically
- No rollback capability
- No migration history
- Not recommended for production

## Recommended: golang-migrate

For production deployments, switch to versioned migrations using golang-migrate.

### Setup

1. **Install migrate tool**:
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

2. **Create migrations directory**:
```bash
mkdir -p backend/migrations
```

3. **Generate initial migration from current schema**:
```bash
# Export current schema
pg_dump -s $DATABASE_URL > schema.sql

# Create initial migration
migrate create -ext sql -dir backend/migrations -seq initial_schema
```

4. **Add migration to Dockerfile**:
```dockerfile
# Add to backend/Dockerfile before CMD
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
COPY migrations ./migrations
CMD ./server -migrate-up && ./server
```

### Usage

**Create new migration**:
```bash
migrate create -ext sql -dir backend/migrations -seq add_user_avatar
```

This creates:
- `000002_add_user_avatar.up.sql`
- `000002_add_user_avatar.down.sql`

**Apply migrations**:
```bash
migrate -path backend/migrations -database "$DATABASE_URL" up
```

**Rollback migration**:
```bash
migrate -path backend/migrations -database "$DATABASE_URL" down 1
```

**Check version**:
```bash
migrate -path backend/migrations -database "$DATABASE_URL" version
```

### Integration with Application

Update [backend/cmd/server/main.go](backend/cmd/server/main.go):

```go
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(databaseURL string) error {
    m, err := migrate.New(
        "file://migrations",
        databaseURL,
    )
    if err != nil {
        return err
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }

    return nil
}

// In main():
if err := runMigrations(cfg.DatabaseURL); err != nil {
    log.Fatalf("Failed to run migrations: %v", err)
}
```

## Migration Workflow

### Development
1. Update model in `backend/internal/models/`
2. Create migration: `migrate create -ext sql -dir backend/migrations -seq describe_change`
3. Write SQL in `.up.sql` file
4. Write rollback SQL in `.down.sql` file
5. Test locally: `migrate -path backend/migrations -database "$DATABASE_URL" up`

### Deployment
1. Migrations run automatically on server startup
2. Railway/Render will run migrations before starting the app
3. If migration fails, deployment stops

### Preview Branches (Neon)
1. Create Neon branch from main (includes all data)
2. Migrations run on preview branch
3. Test changes without affecting production
4. Delete branch when PR is merged

## Example Migrations

### Adding a Column

**up.sql**:
```sql
ALTER TABLE users ADD COLUMN avatar_url VARCHAR(255);
```

**down.sql**:
```sql
ALTER TABLE users DROP COLUMN avatar_url;
```

### Creating a Table

**up.sql**:
```sql
CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
```

**down.sql**:
```sql
DROP TABLE notifications;
```

### Adding an Index

**up.sql**:
```sql
CREATE INDEX idx_messages_thread_created ON messages(thread_id, created_at DESC);
```

**down.sql**:
```sql
DROP INDEX idx_messages_thread_created;
```

## Best Practices

1. **Always test migrations** on preview environment first
2. **Write reversible migrations** when possible
3. **One logical change per migration** for clarity
4. **Never edit applied migrations** - create new ones
5. **Backup before major migrations** (Neon does this automatically)
6. **Test both up and down migrations**
7. **Keep migrations in version control**

## Converting from AutoMigrate

To switch from AutoMigrate to golang-migrate:

1. **Export current schema**:
```bash
pg_dump -s $DATABASE_URL > current_schema.sql
```

2. **Create initial migration**:
```bash
migrate create -ext sql -dir backend/migrations -seq initial_schema
# Copy schema to 000001_initial_schema.up.sql
```

3. **Update code** to use migrate instead of AutoMigrate

4. **Deploy** - golang-migrate tracks versions, won't re-run existing schema

## Neon-Specific Features

### Database Branching
```bash
# Create branch for testing migration
neon branches create --name test-migration --parent main

# Get connection string for test branch
neon connection-string test-migration

# Test migration
migrate -path backend/migrations -database "$TEST_DB_URL" up

# If successful, merge to main
# Delete test branch
neon branches delete test-migration
```

### Point-in-Time Recovery
Neon provides automatic point-in-time recovery:
- Restore to any point in the last 7 days (free tier)
- Useful if migration goes wrong in production

## Troubleshooting

### "Dirty database version"
```bash
# Check current version
migrate -path backend/migrations -database "$DATABASE_URL" version

# Force version (use carefully!)
migrate -path backend/migrations -database "$DATABASE_URL" force VERSION
```

### Migration fails halfway
1. Check error logs
2. Fix migration file
3. Rollback: `migrate down 1`
4. Re-apply: `migrate up`

### Can't rollback AutoMigrate changes
- Use Neon branching to test before applying
- Create manual SQL to reverse changes
- Restore from Neon backup if needed

## Timeline for Production

For immediate production deployment, AutoMigrate is acceptable for MVP. Plan migration to golang-migrate:

- **Week 1**: Continue with AutoMigrate
- **Week 2-3**: Set up golang-migrate, create initial schema migration
- **Week 4**: Switch to golang-migrate in production

This gives you time to deploy quickly while planning proper migration strategy.
