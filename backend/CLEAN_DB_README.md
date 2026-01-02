# Database Clean Script

This script removes conversation data from the database while preserving user accounts and reference data.

## What Gets Deleted

The following tables will be completely cleared:
- `threads` - All conversation threads
- `messages` - All messages
- `tracked_offers` - All tracked offers
- `dealers` - All dealer records

## What Gets Preserved

The following tables are **NOT** affected:
- `users` - All user accounts (preserved)
- `user_preferences` - User vehicle preferences (preserved)
- `gmail_tokens` - Gmail OAuth tokens (preserved)
- `makes` - Vehicle makes (e.g., Toyota, Honda)
- `message_types` - Message type definitions (EMAIL, PHONE)
- `models` - Vehicle models (e.g., Camry, Accord)
- `vehicle_trims` - Vehicle trim specifications

## Usage

### Option 1: Using the Shell Script (Recommended)

```bash
cd backend
./clean-db.sh
```

The script will:
- Use `DATABASE_URL` if set, or
- Prompt for database connection details
- Ask for confirmation before proceeding
- Run the cleanup and show progress

### Option 2: Using psql Directly

```bash
# Using connection string
psql "postgresql://user:password@host:port/database" -f clean-db.sql

# Using connection parameters
psql -h localhost -p 5432 -U postgres -d carbuyer -f clean-db.sql
```

### Option 3: From psql Prompt

```sql
\i clean-db.sql
```

## Environment Variables

You can set these environment variables to avoid prompts:

- `DATABASE_URL` - Full PostgreSQL connection string (takes precedence)
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_NAME` - Database name (default: carbuyer)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (optional, will prompt if not set)

## Example

```bash
# Set environment variables
export DATABASE_URL="postgresql://user:password@localhost:5432/carbuyer"

# Run cleanup
cd backend
./clean-db.sh
```

## Safety

- The script uses a transaction (BEGIN/COMMIT) so if anything fails, all changes are rolled back
- You must explicitly confirm before deletion when using the shell script
- Reference data (makes, models, etc.) is never touched
- The script shows counts before and after cleanup

## After Cleanup

After running the script:
1. All user accounts remain intact (users can still log in)
2. User preferences are preserved
3. All threads and messages will be deleted
4. All tracked offers will be deleted
5. All dealer records will be deleted
6. Reference data (makes, models, message_types, vehicle_trims) remains intact
7. The database schema structure is preserved

