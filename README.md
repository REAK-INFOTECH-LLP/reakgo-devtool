# Reakgo DevTool

Reakgo DevTool is a command-line tool to simplify the setup and management of your Golang projects. It provides functionality for creating databases, generating boilerplate code, and handling database migrations.

## Installation

1. Build the executable file by running:

   ```shell
   go build main.go
This will create an executable file named main in the same directory.


Move the executable file to a directory accessible from anywhere using the following command (you may need sudo privileges):

    sudo mv main /bin

### Getting Started

In any directory where you want to start a new project, follow these steps:

    Initialize the project by creating a database. Run:

    main init

    This command will prompt you for database connection details and create a database.

### Generate the boilerplate code for your Golang project by running:

    main create

This command will create the basic framework for your project, including controllers and models.

### Database Migrations

Reakgo DevTool also supports database migrations. To apply database migrations, follow these steps:

    Ensure you have a migration folder in your project directory. This folder should contain SQL migration files with filenames in the format timestamp.sql.
    like (1695818387.sql)

Run the migration command:

    main migration

This command will execute all SQL migration files in the migration folder that haven't been applied yet. It will also record the applied migrations in the applied_migration folder to prevent them from being executed again.

    Note: Make sure both the migration and applied_migration folders are in the same directory as your controllers and models.

### Changing the Database Schema

If you need to make changes to your database schema, follow these steps:

1. Open the existing migration file in the `migration` folder that corresponds to the change you want to make. These files have names in the format `timestamp.sql`, for example, `1695818387.sql`.

2. Modify the SQL statements in the file to reflect the changes you want to apply to the database schema.

3. Save the changes to the migration file.

4. Run the migration command:

   ```shell
   main migration
Now your project is ready to go with the database setup and boilerplate code generation. You can also manage database migrations effortlessly with Reakgo DevTool.

Happy coding!