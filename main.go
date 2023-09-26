package main

import (
	"archive/zip"
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	dbUser, dbPassword, dbName := promptForDatabaseInfo()

	// Initialize the database and run migrations
	if err := initDB(dbUser, dbPassword, dbName); err != nil {
		return err
	}

	return nil
}

func promptForDatabaseInfo() (string, string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Please enter the database username: ")
	dbUser, _ := reader.ReadString('\n')

	fmt.Print("Please enter the database password: ")
	dbPassword, _ := reader.ReadString('\n')

	fmt.Print("Please enter the database name you want to create: ")
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
