# GobFu (Go Obfuscator)

A Go-based tool for obfuscating sensitive data in relational databases to protect PII (Personally Identifiable Information) while maintaining referential integrity and data structure.

## What is GobFu?

GobFu generates SQL scripts that replace sensitive data in database tables with realistic but fake information. It's designed for creating safe, production-like datasets that can be used for development, testing, and debugging without exposing real customer data.

## Key Features

- **Column-level obfuscation**: Target specific columns containing sensitive data
- **Data type preservation**: Maintains proper data types and formats
- **Configurable data sources**: Use different fake data generators for different column types
- **Referential integrity**: Preserves database relationships
- **PostgreSQL support**: Currently optimized for PostgreSQL databases

## How It Works

1. You define a YAML configuration file specifying which tables and columns contain sensitive data
2. GobFu generates a SQL script that creates temporary tables with fake data
3. The script then updates your database tables with the fake data
4. The result is a database with the same structure but with obfuscated sensitive information

## Installation

### Build from source

```sh
make
```

### Or install it into your $GOPATH

```sh
make install
```

## Usage

### Run directly

```sh
go run cmd/gobfu/main.go -config example-config.yaml -output example.sql 
```

### Or use the built binary

```sh
./bin/gobfu -config example-config.yaml -output example.sql
```

### Or if installed

```sh
gobfu -config example-config.yaml -output example.sql
```

## Configuration

The configuration file uses YAML format to define which tables and columns should be obfuscated and what type of fake data to use.

Example structure:

```yaml
table_name:
  column_name:
    type: data_type
    source: fake_data_source
```

Where:
- `table_name`: The database table containing sensitive data
- `column_name`: The specific column to obfuscate
- `data_type`: The SQL data type (text, jsonb, date, etc.)
- `fake_data_source`: The type of fake data to use (first-name, last-name, email, address, etc.)

[View Example Configuration File](./example-config.yaml)

## Example Output

The tool generates SQL that:
1. Creates temporary tables to store fake data
2. Populates these tables with realistic fake values
3. Updates the target database tables with the fake data
4. Cleans up temporary tables

[View Example Output SQL](./example-output.sql)

## Recommended Workflow

GobFu is designed to be used in a workflow like this:

1. Create a backup/dump of your production database
2. Restore this backup to a secure, isolated environment
3. Run the GobFu-generated SQL script against this database to obfuscate sensitive data
4. Create a new dump of the now-obfuscated database
5. Share this sanitized database dump with developers or use it in non-production environments

This workflow has been successfully implemented using containers in Kubernetes environments.

## Performance Considerations

It is highly recommended to disable database triggers before running obfuscation scripts, especially for audit logs or event triggers. This prevents unwanted side effects such as:

- Significantly longer execution times
- Increased table sizes from trigger-generated data
- Unwanted notifications or side effects

### Disabling Triggers (PostgreSQL)

```sql
DO $$
    DECLARE
        r RECORD;
    BEGIN
        FOR r IN (SELECT tgrelid::regclass AS table_name
                  FROM pg_trigger
                  WHERE NOT tgisinternal)
            LOOP
                EXECUTE 'ALTER TABLE ' || r.table_name || ' DISABLE TRIGGER ALL';
            END LOOP;
    END $$;
```

### Re-enabling Triggers (PostgreSQL)

```sql
DO $$
    DECLARE
        r RECORD;
    BEGIN
        FOR r IN (SELECT tgrelid::regclass AS table_name
                  FROM pg_trigger
                  WHERE NOT tgisinternal)
            LOOP
                EXECUTE 'ALTER TABLE ' || r.table_name || ' ENABLE TRIGGER ALL';
            END LOOP;
    END $$;
```

## Future Enhancements

- Support for additional database dialects beyond PostgreSQL
- Docker-based automation scripts for the complete workflow
- Custom data source definitions
- More sophisticated data transformation options
