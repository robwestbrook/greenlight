# Greenlight
This is my repository created while following the book ***Let's Go Further*** by **Alex Edwards**.

In the book, a movie API is built. I changed this to an event API, which I hope to use as the basis for a calender API and app later.

The book uses **Postgres** as the database. This repository, instead, uses **Sqlite**. There are some modifications and additions made to the book's code to accomodate this change. These are documented within the code.

### Using Migrate CLI
The command-line statement to use for migration:

    migrate -path=./migrations -database=sqlite3://greenlight.db -verbose up