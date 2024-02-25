# Greenlight
This is my repository created while following the book ***Let's Go Further*** by **Alex Edwards**.

In the book, a movie API is built. I changed this to an event API, which I hope to use as the basis for a calender API and app later.

### Using SQLite
The book uses **Postgres** as the database. This repository, instead, uses **Sqlite**. There are quite a few modifications and additions made to the book's code to accomodate this change. These are documented within the code.

### .env
I am also using the **godotenv** package for applocation settings. To install this package:

    go get github.com/joho/godotenv

### Using Migrate CLI
The command-line statements to use for database migration.

To create a migration:

     migrate create -seq -ext=.sql -dir=./migrations <migration name>

To run a migration (use up or down):

    migrate -path=./migrations -database=sqlite3://greenlight.db -verbose up

### Versioning
The application uses a combination of Git tags and
Git commit hashes for versioning. To add a tag using Git:

    git tag v<version-number>

The version output will be in this format:
    
    {tag}-{number of additional commits}-g{abbreviated commit hash}