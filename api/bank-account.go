package api

import (
	"database/sql"
	"errors"
	"fmt"

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
	Create(account *BankAccount) error
	Delete(id int) error
	Withdraw(id, amount int) (*BankAccount, error)
	Deposit(id, amount int) (*BankAccount, error)
	Transfer(amount, fromAccID, toAccID int) error
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

func (s *AccountServiceMySQL) Create(acc *BankAccount) error {
	insertStmt := "INSERT INTO ACCOUNT(ID, USER_ID, ACCOUNT_NUMBER, NAME, BALANCE) VALUES(?,?,?,?,?)"
	if result, err := s.DB.Exec(insertStmt, acc.ID, acc.UserID, acc.AccountNumber, acc.Name, acc.Balance); err != nil {
		return err
	} else {
		i, _ := result.LastInsertId()
		acc.ID = int(i)
		fmt.Printf("Insert an account [%s] completed with ID:%s\n", acc.AccountNumber, i)
		return nil
	}
}

func (s *AccountServiceMySQL) Delete(id int) error {
	stmt := "DELETE FROM BANK_ACCOUNT where id = ?"
	if _, err := s.DB.Exec(stmt, id); err != nil {
		return err
	} else {
		fmt.Printf("Deletion an account [%d] completed\n", id)
		return nil
	}
}

func (s *AccountServiceMySQL) Withdraw(id, amount int) (*BankAccount, error) {
	stmt := "UPDATE BANK_ACCOUNT SET BALANCE = BALANCE - ?1 WHERE ID = ?2 AND BALANCE > ?1"
	if res, err := s.DB.Exec(stmt, amount, id); err != nil {
		return nil, err
	} else {
		updated, err := res.RowsAffected()
		if err != nil {
			return nil, err
		} else if updated < 1 {
			return nil, errors.New("Bank account not found or insufficient amount")
		} else {
			fmt.Printf("Withdraw an account [%d] completed\n", id)
			return s.getAccountByID(id)
		}
	}
}

func (s *AccountServiceMySQL) Deposit(id, amount int) (*BankAccount, error) {
	stmt := "UPDATE BANK_ACCOUNT SET BALANCE = BALANCE + ? WHERE ID = ?"
	if res, err := s.DB.Exec(stmt, amount, id); err != nil {
		return nil, err
	} else {
		updated, err := res.RowsAffected()
		if err != nil {
			return nil, err
		} else if updated < 1 {
			return nil, errors.New("Bank account not found")
		} else {
			fmt.Printf("Deposit an account [%d] completed\n", id)
			return s.getAccountByID(id)
		}
	}
}

func (s *AccountServiceMySQL) getAccountByID(id int) (*BankAccount, error) {
	stmt := "SELECT ID, USER_ID, ACCOUNT_NUMBER, NAME, BALANCE FROM BANK_ACCOUNT WHERE ID = ?"
	row := s.DB.QueryRow(stmt, id)

	var acc BankAccount
	if err := row.Scan(&acc.ID, &acc.UserID, &acc.AccountNumber, &acc.Name, &acc.Balance); err != nil {
		return nil, err
	} else {
		return &acc, nil
	}
}

func (s *AccountServiceMySQL) Transfer(amount, fromAccID, toAccID int) error {
	if tx, err := s.DB.Begin(); err != nil {
		return err
	} else {
		withdrawStmt := "UPDATE BANK_ACCOUNT SET BALANCE = BALANCE - ?1 WHERE ID = ?2 AND BALANCE > ?1"

		res, err := tx.Exec(withdrawStmt, amount, fromAccID)
		if err != nil {
			tx.Rollback()
			return err
		}
		updated, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return err
		}
		if updated > 0 {
			depositStmt := "UPDATE BANK_ACCOUNT SET BALANCE = BALANCE + ? WHERE ID = ?"
			res, err = tx.Exec(depositStmt, amount, toAccID)
			if err != nil {
				tx.Rollback()
				return err
			}
			updated, err := res.RowsAffected()
			if err != nil {
				tx.Rollback()
				return err
			}
			if updated > 0 {
				return nil
			} else {
				return errors.New("To bank account not found")
			}
		} else {
			return errors.New("From bank account not found or insufficient amount")
		}
	}

}
