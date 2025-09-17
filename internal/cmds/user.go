package cmds

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/urfave/cli/v2"
	"golang.org/x/term"

	"famstack/internal/auth"
	"famstack/internal/config"
	"famstack/internal/database"
	"famstack/internal/encryption"
)

// UserCommand returns the user management command configuration
func UserCommand() *cli.Command {
	return &cli.Command{
		Name:    "user",
		Aliases: []string{"u"},
		Usage:   "User management commands",
		Subcommands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create a new user",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "email",
						Usage: "User email address",
					},
					&cli.StringFlag{
						Name:  "first-name",
						Usage: "User first name",
					},
					&cli.StringFlag{
						Name:  "last-name",
						Usage: "User last name",
					},
					&cli.StringFlag{
						Name:  "role",
						Usage: "User role (shared, user, admin)",
						Value: "user",
					},
					&cli.StringFlag{
						Name:  "family-id",
						Usage: "Family ID to associate with user",
						Value: "fam1",
					},
					&cli.StringFlag{
						Name:  "password",
						Usage: "User password (WARNING: visible in process list)",
					},
					&cli.StringFlag{
						Name:  "db",
						Value: "famstack.db",
						Usage: "Database file path",
					},
				},
				Action: createUser,
			},
			{
				Name:  "list",
				Usage: "List all users",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "db",
						Value: "famstack.db",
						Usage: "Database file path",
					},
				},
				Action: listUsers,
			},
			{
				Name:  "delete",
				Usage: "Delete a user",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "email",
						Usage:    "User email address",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "db",
						Value: "famstack.db",
						Usage: "Database file path",
					},
				},
				Action: deleteUser,
			},
			{
				Name:  "reset-password",
				Usage: "Reset a user's password",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "email",
						Usage:    "User email address",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "password",
						Usage: "New password (WARNING: visible in process list)",
					},
					&cli.StringFlag{
						Name:  "db",
						Value: "famstack.db",
						Usage: "Database file path",
					},
				},
				Action: resetPassword,
			},
		},
	}
}

func createUser(ctx *cli.Context) error {
	dbPath := ctx.String("db")

	// Initialize database
	db, err := database.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Initialize encryption service
	encryptionConfig := config.DefaultEncryptionSettings()
	encryptionService, err := encryption.NewService(*encryptionConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize encryption service: %w", err)
	}

	// Initialize auth service
	authService := auth.NewService(db.DB, encryptionService, "famstack")

	// Get user details
	email := ctx.String("email")
	if email == "" {
		fmt.Print("Enter email: ")
		reader := bufio.NewReader(os.Stdin)
		var readErr error
		email, readErr = reader.ReadString('\n')
		if readErr != nil {
			return fmt.Errorf("failed to read email: %w", readErr)
		}
		email = strings.TrimSpace(email)
	}

	firstName := ctx.String("first-name")
	if firstName == "" {
		fmt.Print("Enter first name: ")
		reader := bufio.NewReader(os.Stdin)
		var readErr error
		firstName, readErr = reader.ReadString('\n')
		if readErr != nil {
			return fmt.Errorf("failed to read first name: %w", readErr)
		}
		firstName = strings.TrimSpace(firstName)
	}

	lastName := ctx.String("last-name")
	if lastName == "" {
		fmt.Print("Enter last name: ")
		reader := bufio.NewReader(os.Stdin)
		var readErr error
		lastName, readErr = reader.ReadString('\n')
		if readErr != nil {
			return fmt.Errorf("failed to read last name: %w", readErr)
		}
		lastName = strings.TrimSpace(lastName)
	}

	// Get password
	password := ctx.String("password")
	if password == "" {
		// Get password securely
		fmt.Print("Enter password: ")
		passwordBytes, pwdErr := term.ReadPassword(int(syscall.Stdin))
		if pwdErr != nil {
			return fmt.Errorf("failed to read password: %w", pwdErr)
		}
		password = string(passwordBytes)
		fmt.Println() // New line after password input

		fmt.Print("Confirm password: ")
		confirmBytes, confirmErr := term.ReadPassword(int(syscall.Stdin))
		if confirmErr != nil {
			return fmt.Errorf("failed to read password confirmation: %w", confirmErr)
		}
		confirm := string(confirmBytes)
		fmt.Println() // New line after password input

		if password != confirm {
			return fmt.Errorf("passwords do not match")
		}
	}

	role := ctx.String("role")
	familyID := ctx.String("family-id")

	// Create user
	req := &auth.CreateUserRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		Role:      auth.Role(role),
		FamilyID:  familyID,
	}

	user, err := authService.CreateUser(req)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Printf("✅ User created successfully!\n")
	fmt.Printf("ID: %s\n", user.ID)
	if user.Email != nil {
		fmt.Printf("Email: %s\n", *user.Email)
	}
	fmt.Printf("Name: %s\n", user.Name)
	if user.Role != nil {
		fmt.Printf("Role: %s\n", *user.Role)
	}
	fmt.Printf("Family ID: %s\n", user.FamilyID)

	return nil
}

func listUsers(ctx *cli.Context) error {
	dbPath := ctx.String("db")

	// Initialize database
	db, err := database.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Query family members with auth info
	rows, err := db.Query(`
		SELECT id, email, name, member_type, role, family_id, created_at
		FROM family_members
		WHERE password_hash IS NOT NULL
		ORDER BY created_at DESC
	`)
	if err != nil {
		return fmt.Errorf("failed to query family members: %w", err)
	}
	defer rows.Close()

	fmt.Printf("%-8s %-25s %-20s %-10s %-8s %-8s %-20s\n",
		"ID", "Email", "Name", "Type", "Role", "Family", "Created At")
	fmt.Println(strings.Repeat("-", 100))

	for rows.Next() {
		var id, familyID, createdAt, name, memberType string
		var email, role sql.NullString

		err := rows.Scan(&id, &email, &name, &memberType, &role, &familyID, &createdAt)
		if err != nil {
			return fmt.Errorf("failed to scan family member: %w", err)
		}

		// Handle nullable fields and truncate if needed
		emailStr := ""
		if email.Valid {
			emailStr = email.String
			if len(emailStr) > 25 {
				emailStr = emailStr[:22] + "..."
			}
		}

		roleStr := ""
		if role.Valid {
			roleStr = role.String
		}

		// Truncate other fields if too long
		if len(id) > 8 {
			id = id[:8]
		}
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		fmt.Printf("%-8s %-25s %-20s %-10s %-8s %-8s %-20s\n",
			id, emailStr, name, memberType, roleStr, familyID, createdAt)
	}

	return nil
}

func deleteUser(ctx *cli.Context) error {
	dbPath := ctx.String("db")
	email := ctx.String("email")

	// Initialize database
	db, err := database.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete user with email '%s'? (y/N): ", email)
	reader := bufio.NewReader(os.Stdin)
	response, readErr := reader.ReadString('\n')
	if readErr != nil {
		return fmt.Errorf("failed to read confirmation: %w", readErr)
	}
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	// Delete family member
	result, err := db.Exec("DELETE FROM family_members WHERE email = ?", email)
	if err != nil {
		return fmt.Errorf("failed to delete family member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		fmt.Printf("❌ No user found with email '%s'\n", email)
		return nil
	}

	fmt.Printf("✅ User with email '%s' deleted successfully\n", email)
	return nil
}

func resetPassword(ctx *cli.Context) error {
	dbPath := ctx.String("db")
	email := ctx.String("email")

	// Initialize database
	db, err := database.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Check if family member exists and has auth
	var userID string
	err = db.QueryRow("SELECT id FROM family_members WHERE email = ? AND password_hash IS NOT NULL", email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("family member with email '%s' not found or cannot authenticate", email)
		}
		return fmt.Errorf("failed to query family member: %w", err)
	}

	// Get new password
	password := ctx.String("password")
	if password == "" {
		// Get password securely
		fmt.Print("Enter new password: ")
		passwordBytes, pwdErr := term.ReadPassword(int(syscall.Stdin))
		if pwdErr != nil {
			return fmt.Errorf("failed to read password: %w", pwdErr)
		}
		password = string(passwordBytes)
		fmt.Println() // New line after password input

		fmt.Print("Confirm new password: ")
		confirmBytes, confirmErr := term.ReadPassword(int(syscall.Stdin))
		if confirmErr != nil {
			return fmt.Errorf("failed to read password confirmation: %w", confirmErr)
		}
		confirm := string(confirmBytes)
		fmt.Println() // New line after password input

		if password != confirm {
			return fmt.Errorf("passwords do not match")
		}
	}

	// Hash the new password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update family member's password
	result, err := db.Exec(
		"UPDATE family_members SET password_hash = ? WHERE email = ?",
		hashedPassword, email,
	)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with email '%s'", email)
	}

	fmt.Printf("✅ Password reset successfully for user '%s'\n", email)
	return nil
}
