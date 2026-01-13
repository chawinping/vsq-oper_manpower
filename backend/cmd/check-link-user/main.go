package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run main.go <username>")
	}
	username := os.Args[1]

	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "vsq_user")
	dbPassword := getEnv("DB_PASSWORD", "vsq_password")
	dbName := getEnv("DB_NAME", "vsq_manpower")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Check if user exists
	var userID string
	var branchID sql.NullString
	var roleName string
	err = db.QueryRow(`
		SELECT u.id, u.branch_id, r.name as role_name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1
	`, username).Scan(&userID, &branchID, &roleName)

	if err == sql.ErrNoRows {
		log.Fatalf("❌ User '%s' not found in database", username)
	}
	if err != nil {
		log.Fatalf("❌ Error querying user: %v", err)
	}

	fmt.Printf("User: %s\n", username)
	fmt.Printf("User ID: %s\n", userID)
	fmt.Printf("Role: %s\n", roleName)

	if branchID.Valid {
		// Get branch details
		var branchCode, branchName string
		err = db.QueryRow(`
			SELECT code, name FROM branches WHERE id = $1
		`, branchID.String).Scan(&branchCode, &branchName)
		if err != nil {
			log.Fatalf("❌ Error querying branch: %v", err)
		}
		fmt.Printf("✅ Branch ID: %s\n", branchID.String)
		fmt.Printf("✅ Branch Code: %s\n", branchCode)
		fmt.Printf("✅ Branch Name: %s\n", branchName)
	} else {
		fmt.Printf("❌ Branch ID: NOT SET\n")

		// Try to determine expected branch code from username
		expectedBranchCode := ""
		if len(username) >= 3 {
			if len(username) >= 5 && username[len(username)-4:] == "amgr" {
				expectedBranchCode = username[:len(username)-4]
			} else if len(username) >= 4 && username[len(username)-3:] == "mgr" {
				expectedBranchCode = username[:len(username)-3]
			}
		}

		if expectedBranchCode != "" {
			expectedBranchCode = strings.ToUpper(expectedBranchCode)
			fmt.Printf("Expected branch code (from username): %s\n", expectedBranchCode)

			// Check if branch exists
			var branchIDToLink string
			var branchName string
			err = db.QueryRow(`
				SELECT id, name FROM branches WHERE code = $1
			`, expectedBranchCode).Scan(&branchIDToLink, &branchName)

			if err == sql.ErrNoRows {
				fmt.Printf("❌ Branch with code '%s' not found in database\n", expectedBranchCode)
				os.Exit(1)
			}
			if err != nil {
				log.Fatalf("❌ Error querying branch: %v", err)
			}

			// Ask for confirmation if not in auto-link mode
			if len(os.Args) < 3 || os.Args[2] != "--link" {
				fmt.Printf("\nTo link this user to branch %s (%s), run:\n", expectedBranchCode, branchName)
				fmt.Printf("  go run main.go %s --link\n", username)
				os.Exit(0)
			}

			// Link the user
			_, err = db.Exec("UPDATE users SET branch_id = $1 WHERE id = $2", branchIDToLink, userID)
			if err != nil {
				log.Fatalf("❌ Failed to link user to branch: %v", err)
			}

			fmt.Printf("✅ Successfully linked user '%s' to branch %s (%s)\n", username, expectedBranchCode, branchName)
		} else {
			fmt.Printf("❌ Could not determine expected branch code from username pattern\n")
			fmt.Printf("Username should follow pattern: {branchcode}mgr or {branchcode}amgr\n")
			os.Exit(1)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

