package main

import (
	"archive/zip"
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sort"

	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	// Check if any command-line arguments were provided
	if len(os.Args) > 1 {
		// If the first argument is "init," create the .reakgo file and return
		if os.Args[1] == "init" {
			err := initReakgoFile()
			if err != nil {
				fmt.Println("Error initializing .reakgo file:", err)
			} else {
				fmt.Println("ready for use.")
			}
			return
		} else if os.Args[1] == "create" {
			// Check if .reakgo file exists
			if _, err := os.Stat(".reakgo"); os.IsNotExist(err) {
				fmt.Println("Error:Please run 'init' first.if you already run that then please check the directory.")
				return
			}
			err := boilerPlateCreate()
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("file created and ready for use")
			}
			return

		} else if os.Args[1] == "migration" {
			// Automatically run migrations
			if err := runMigrations("./migrations"); err != nil {
				fmt.Println("Migration Error:", err)
				return
			}

		} else {
			log.Println("please correct the command")
		}
	} else {
		log.Println("Too many arguments")
	}
	return

}

func initReakgoFile() error {
	// Create a .reakgo file in the current directory
	err := os.WriteFile(".reakgo", []byte("Reakgo configuration"), 0644)
	if err != nil {
		return err
	}
	// Prompt the user for database connection details
	dbUser, dbPassword, dbName := promptForDatabaseInfo("")

	// Initialize the database and run migrations
	if err := initDB(dbUser, dbPassword, dbName); err != nil {
		return err
	}

	return nil
}

func promptForDatabaseInfo(check string) (string, string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Please enter the database username: ")
	dbUser, _ := reader.ReadString('\n')

	fmt.Print("Please enter the database password: ")
	dbPassword, _ := reader.ReadString('\n')
	if check == "migration" {
		fmt.Print("Please enter the database name: ")
	} else {
		fmt.Print("Please enter the database name you want to create: ")
	}
	dbName, _ := reader.ReadString('\n')

	// Remove trailing newline characters
	dbUser = strings.TrimSpace(dbUser)
	dbPassword = strings.TrimSpace(dbPassword)
	dbName = strings.TrimSpace(dbName)

	return dbUser, dbPassword, dbName
}

func initDB(dbUser, dbPassword, dbName string) error {
	// Initialize the database connection
	dbURL := fmt.Sprintf("%s:%s@/", dbUser, dbPassword)
	var err error
	db, err = sql.Open("mysql", dbURL) // Change to your preferred database driver
	if err != nil {
		return err
	}

	// Check the database connection
	if err := db.Ping(); err != nil {
		return err
	}
	// Create the database
	_, err = db.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		return err
	}

	err = importSQLFile(dbName, "testdatabase.sql", dbUser, dbPassword) // Replace with the actual SQL file name
	if err != nil {
		fmt.Println("Error importing SQL file:", err)
		return err
	}

	log.Println("Database initialized successfully.")

	return nil
}
func importSQLFile(dbName, sqlFileName, dbUser, dbPassword string) error {
	// Use the `mysql` command-line tool to import the SQL file into the database
	cmd := exec.Command("mysql", dbName, fmt.Sprintf("--user=%s", dbUser), fmt.Sprintf("--password=%s", dbPassword), "-e", fmt.Sprintf("source %s", sqlFileName))
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func boilerPlateCreate() error {
	// Define the Git repository URL and ZIP archive URL
	repoURL := "https://github.com/REAK-INFOTECH-LLP/reakgo"
	zipURL := repoURL + "/archive/master.zip" // Replace "master" with the branch or tag you want to download

	// Define the name of the ZIP file
	zipFileName := "repo.zip"

	// Create a new file to save the ZIP archive
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		fmt.Println("Error creating ZIP file:", err)
		return err
	}
	defer zipFile.Close()

	// Send an HTTP GET request to the ZIP archive URL
	resp, err := http.Get(zipURL)
	if err != nil {
		fmt.Println("Error sending HTTP GET request:", err)
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		status_err := fmt.Sprintf("HTTP GET request failed with status: %s\n", resp.Status)
		return errors.New(status_err)
	}

	// Copy the response body (ZIP archive) to the ZIP file
	_, err = io.Copy(zipFile, resp.Body)
	if err != nil {
		fmt.Println("Error copying ZIP archive to file:", err)
		return err
	}
	// Extract the ZIP archive in the same directory
	err = unzip(zipFileName, ".")
	if err != nil {
		fmt.Println("Error extracting ZIP archive:", err)
		return err
	}

	// Delete the ZIP file
	err = os.Remove(zipFileName)
	if err != nil {
		fmt.Println("Error deleting ZIP file:", err)
		return err
	}

	return nil
}

func unzip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		filePath := filepath.Join(dest, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			return err
		}

		outFile, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		_, err = io.Copy(outFile, fileReader)
		if err != nil {
			return err
		}
	}

	return nil
}
func runMigrations(migrationsDir string) error {

	// Prompt the user for database connection details
	dbUser, dbPassword, dbName := promptForDatabaseInfo("migration")
	// Initialize the database connection
	// dbURL := fmt.Sprintf("%s:%s@/", dbUser, dbPassword)
	dbURL := fmt.Sprintf("%s:%s@/%s", dbUser, dbPassword, dbName)
	var err error
	db, err = sql.Open("mysql", dbURL) // Change to your preferred database driver
	if err != nil {
		return err
	}

	// Check the database connection
	if err := db.Ping(); err != nil {
		return err
	}

	// Get the list of applied migrations
	appliedMigrations, err := getAppliedMigrations()
	if err != nil {
		log.Println(err)
	}

	// List migration files in the directory
	migrationFiles, err := listMigrationFiles(migrationsDir)
	if err != nil {
		log.Println(err)
	}

	// Sort migration files
	sort.Strings(migrationFiles)

	// Apply pending migrations
	for _, migrationFile := range migrationFiles {
		if !stringSliceContains(appliedMigrations, migrationFile) {
			// Read the migration SQL from the file
			migrationSQL, err := readFileContents(filepath.Join(migrationsDir, migrationFile))
			if err != nil {
				log.Println(err)
				return err
			}
			// Begin a transaction
			tx, err := db.Begin()
			if err != nil {
				fmt.Println("Error starting transaction:", err)
				return err
			}
			// Split SQL statements into individual statements
			statements := strings.Split(migrationSQL, ";")
			// Execute each SQL statement
			for _, statement := range statements {
				// Trim leading and trailing whitespace
				statement = strings.TrimSpace(statement)
				// Skip empty statements
				if statement == "" {
					continue
				}
				statement = statement + ";"
				// Execute the SQL statement
				_, err := tx.Exec(statement)
				if err != nil {
					fmt.Println("Error executing SQL statement:", err)
					tx.Rollback()
					return err
				}
			}
			err = tx.Commit()
			if err != nil {
				fmt.Println("Error commiting SQL statement:", err)
				return err
			}
			// Record the applied migration in a file
			err = recordMigration(migrationFile)
			if err != nil {
				log.Println(err)
				return err
			}

			fmt.Printf("Applied migration: %s\n", migrationFile)
		}
	}

	fmt.Println("All pending migrations applied successfully.")
	return nil
}

// Get a list of applied migrations based on recorded files
func getAppliedMigrations() ([]string, error) {
	var appliedMigrations []string
	// Define a directory to store applied migration files
	appliedDir := "./applied_migrations" // Change to your desired directory

	// Create the directory if it doesn't exist
	if _, err := os.Stat(appliedDir); os.IsNotExist(err) {
		err := os.Mkdir(appliedDir, os.ModePerm)
		if err != nil {
			return appliedMigrations, err
		}
	}

	// List applied migration files
	files, err := ioutil.ReadDir(appliedDir)
	if err != nil {
		return appliedMigrations, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			appliedMigrations = append(appliedMigrations, file.Name())
		}
	}

	return appliedMigrations, nil
}

// Record a migration as applied in a file
func recordMigration(name string) error {
	// Define a directory to store applied migration files
	appliedDir := "./applied_migrations" // Change to your desired directory

	// Create the directory if it doesn't exist
	if _, err := os.Stat(appliedDir); os.IsNotExist(err) {
		err := os.Mkdir(appliedDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Create a new migration file with a unique name based on a timestamp
	filePath := filepath.Join(appliedDir, name)

	// Create and write the migration file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

// List migration files in the directory
func listMigrationFiles(dirPath string) ([]string, error) {
	var migrationFiles []string
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return migrationFiles, err
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	return migrationFiles, nil
}

// Read file contents
func readFileContents(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Check if a string exists in a slice of strings
func stringSliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
