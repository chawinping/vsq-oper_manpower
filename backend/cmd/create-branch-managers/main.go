package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
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

	// Ensure branch_id column exists in users table
	_, err = db.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'users' AND column_name = 'branch_id'
			) THEN
				ALTER TABLE users ADD COLUMN branch_id UUID REFERENCES branches(id);
			END IF;
		END $$;
	`)
	if err != nil {
		log.Fatalf("Failed to add branch_id column: %v", err)
	}

	// Get branch manager role ID
	var branchManagerRoleID string
	err = db.QueryRow("SELECT id FROM roles WHERE name = 'branch_manager'").Scan(&branchManagerRoleID)
	if err != nil {
		log.Fatalf("Failed to get branch_manager role: %v", err)
	}

	// Get all branches
	branches, err := getBranches(db)
	if err != nil {
		log.Fatalf("Failed to get branches: %v", err)
	}

	// Create users for each branch
	created := 0
	skipped := 0
	errors := 0

	for _, branch := range branches {
		branchCode := strings.ToLower(branch.code)
		
		// Create two users per branch: branchcode+mgr and branchcode+amgr
		usernames := []string{
			branchCode + "mgr",   // Branch Manager
			branchCode + "amgr",  // Assistant Branch Manager
		}

		for _, username := range usernames {
			// Check if user already exists
			var existingID string
			var existingBranchID sql.NullString
			err := db.QueryRow("SELECT id, branch_id FROM users WHERE username = $1", username).Scan(&existingID, &existingBranchID)
			if err == nil {
				// User exists - check if branch_id needs to be set
				if !existingBranchID.Valid {
					// Update existing user with branch_id
					_, err = db.Exec("UPDATE users SET branch_id = $1 WHERE id = $2", branch.id, existingID)
					if err != nil {
						log.Printf("❌ Failed to update branch_id for user %s: %v\n", username, err)
						errors++
						continue
					}
					fmt.Printf("✅ Updated user %s: linked to branch %s (%s)\n", username, branch.name, branch.code)
					created++
				} else {
					fmt.Printf("⏭️  User %s already exists with branch_id, skipping...\n", username)
					skipped++
				}
				continue
			} else if err != sql.ErrNoRows {
				log.Printf("❌ Error checking user %s: %v\n", username, err)
				errors++
				continue
			}

			// Create user
			password := username // Default password same as username
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("❌ Failed to hash password for %s: %v\n", username, err)
				errors++
				continue
			}

			email := fmt.Sprintf("%s@vsq.local", username)
			userID := uuid.New()

			// Get branch ID for this user based on branch code
			var branchID string
			err = db.QueryRow("SELECT id FROM branches WHERE code = $1", branch.code).Scan(&branchID)
			if err != nil {
				log.Printf("❌ Failed to get branch ID for %s: %v\n", branch.code, err)
				errors++
				continue
			}

			_, err = db.Exec(`
				INSERT INTO users (id, username, email, password_hash, role_id, branch_id)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, userID, username, email, string(passwordHash), branchManagerRoleID, branchID)
			
			if err != nil {
				log.Printf("❌ Failed to create user %s: %v\n", username, err)
				errors++
				continue
			}

			if err != nil {
				log.Printf("❌ Failed to create user %s: %v\n", username, err)
				errors++
				continue
			}

			fmt.Printf("✅ Created user: %s (password: %s) for branch %s (%s)\n", username, password, branch.name, branch.code)
			created++
		}
	}

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("Summary:\n")
	fmt.Printf("  ✅ Created: %d users\n", created)
	fmt.Printf("  ⏭️  Skipped: %d users (already exist)\n", skipped)
	fmt.Printf("  ❌ Errors: %d\n", errors)
	fmt.Println(strings.Repeat("=", 60))

	if errors > 0 {
		os.Exit(1)
	}
}

type branch struct {
	id   string
	code string
	name string
}

func getBranches(db *sql.DB) ([]branch, error) {
	rows, err := db.Query("SELECT id, code, name FROM branches ORDER BY code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []branch
	for rows.Next() {
		var b branch
		if err := rows.Scan(&b.id, &b.code, &b.name); err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	return branches, rows.Err()
}


func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

