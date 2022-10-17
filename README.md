# Secure HTTP Server

This an example project that is all about security. There will be TLS, authentication via JWTs, Password hashing and more.

We are going to be creating a database

- Delete all entries from a table using `truncate {TABLE_NAME}`

Soft delete: [link](https://evilmartians.com/chronicles/soft-deletion-with-postgresql-but-with-logic-on-the-database)

- [Load env data](https://stackoverflow.com/questions/19331497/set-environment-variables-from-file-of-key-value-pairs)

```bash
export $(grep -v '^#' .env | xargs -d '\n')
```

- Quotes for [bash](https://unix.stackexchange.com/questions/443989/whats-the-right-way-to-quote-command-arg)
