# sqlstore

SQL session store for the Gorilla web toolkit.

Currently only supports the PostgreSQL dialect.

Requires a `sessions` table:

```sql
CREATE TABLE sessions (id varchar(100) PRIMARY KEY, data text NOT NULL);
```
