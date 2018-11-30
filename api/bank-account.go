package api

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type BankAccount struct {
	ID            int    `json:"id"`
	UserID        int    `json:"user_id"`
	AccountNumber string `json:"account_number" binding:"required"`
	Name          string `json:"name"`
	Balance       int    `json:"amount"`
}

type AccountService interface {
	List(userID int) ([]BankAccount, error)
}

type AccountServiceMySQL struct {
	DB *sql.DB
}

func (s *AccountServiceMySQL) List(userID int) ([]BankAccount, error) {
	accounts := make([]BankAccount, 0)
	queryStmt := "SELECT ID, USER_ID, ACCOUNT_NUMBER, NAME, BALANCE FROM BANK_ACCOUNT ORDER BY ID"
	if rows, err := s.DB.Query(queryStmt); err != nil {
		return nil, err
	} else {
		for rows.Next() {
			var acc BankAccount
			rows.Scan(&acc.ID, &acc.UserID, &acc.AccountNumber, &acc.Name, &acc.Balance)
			accounts = append(accounts, acc)
		}
		return accounts, nil
	}
}
