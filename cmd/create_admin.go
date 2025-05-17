package cmd

import (
	"bufio"
	domainuser "carowebapp/core/internal/domain/user"
	"fmt"
	"os"
	"strings"

	"carowebapp/core/internal/features/auth"
	db "carowebapp/core/internal/infrastructure/db"
	"carowebapp/core/internal/infrastructure/email"
	"carowebapp/core/internal/infrastructure/logger"

	"github.com/spf13/cobra"
)

// CreateAdminCmd is the command to create an admin user
var CreateAdminCmd = &cobra.Command{
	Use:   "create-admin",
	Short: "Create an admin user",
	Run: func(_ *cobra.Command, _ []string) { // Используем _ вместо cmd и args
		logger.Init(false)
		dbConn := db.InitDB()
		repo := auth.NewSQLXRepository(dbConn)
		sender := email.NewMailer(logger.Log)
		service := auth.NewService(repo, sender)

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Email: ")
		emailInput, _ := reader.ReadString('\n')
		email := strings.TrimSpace(emailInput)

		fmt.Print("Password: ")
		passwordInput, _ := reader.ReadString('\n')
		password := strings.TrimSpace(passwordInput)

		user, err := service.RegisterUser(email, password, domainuser.RoleAdmin, true)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		fmt.Println("Admin user created:", user.Email)
	},
}
