package api

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"`
}

type UserService interface {
	Insert(u *User) error
	Update(u *User) error
	Delete(id int) error
	FindByID(id int) (*User, error)
	All() ([]User, error)
}

type UserServiceMySQL struct {
	DB *sql.DB
}

func (s *UserServiceMySQL) Insert(u *User) error {
	insertStmt := "INSERT INTO USER(first_name, last_name) VALUES(?,?)"
	if result, err := s.DB.Exec(insertStmt, u.FirstName, u.LastName); err != nil {
		return err
	} else {
		i, _ := result.LastInsertId()
		u.ID = int(i)
		fmt.Printf("Insert a user [%s] completed with ID:%s\n", u.FirstName, i)
		return nil
	}
}

func (s *UserServiceMySQL) Update(u *User) error {
	stmt := "UPDATE USERS SET first_name = ?, last_name = ? WHERE ID = ?"
	if _, err := s.DB.Exec(stmt, u.FirstName, u.LastName, u.ID); err != nil {
		return err
	} else {
		fmt.Printf("Update a user [%s] completed\n", u.FirstName)
		return nil
	}
}

func (s *UserServiceMySQL) Delete(id int) error {
	stmt := "DELETE FROM USER where id = ?"
	if _, err := s.DB.Exec(stmt, id); err != nil {
		return err
	} else {
		fmt.Printf("Deletion a user [%d] completed\n", id)
		return nil
	}
}

func (s *UserServiceMySQL) FindByID(n int) (*User, error) {
	stmt := "SELECT ID, FIRST_NAME, LAST_NAME FROM USER where id = ?"
	row := s.DB.QueryRow(stmt, n)

	var user User
	if err := row.Scan(&user.ID, &user.FirstName, &user.LastName); err != nil {
		return nil, err
	} else {
		return &user, nil
	}
}

func (s *UserServiceMySQL) All() ([]User, error) {
	users := make([]User, 0)
	queryStmt := "SELECT ID, FIRST_NAME, LAST_NAME FROM USER ORDER BY ID"
	if rows, err := s.DB.Query(queryStmt); err != nil {
		return nil, err
	} else {
		for rows.Next() {
			var user User
			rows.Scan(&user.ID, &user.FirstName, &user.LastName)
			users = append(users, user)
		}
		return users, nil
	}
}

/*
func Insert(u *User) error {
	db := getDB()
	defer db.Close()

	insertStmt := "INSERT INTO USERS(first_name, last_name, email) VALUES(?,?,?)"
	if result, err := db.Exec(insertStmt, u.FirstName, u.LastName, u.Email); err != nil {
		handleError(err)
		return err
	} else {
		i, _ := result.LastInsertId()
		fmt.Println("Insert completed with ", i)
		return nil
	}
}

func Update(u *User) error {
	db := getDB()
	defer db.Close()

	stmt := "UPDATE USERS SET first_name = ?, last_name = ?, email = ? WHERE ID = ?"
	if _, err := db.Exec(stmt, u.FirstName, u.LastName, u.Email, u.ID); err != nil {
		handleError(err)
		return err
	} else {
		fmt.Println("Update completed")
		return nil
	}
}

func Last() (*User, error) {
	db := getDB()
	defer db.Close()
	stmt := "SELECT ID, FIRST_NAME, LAST_NAME, EMAIL FROM USERS ORDER BY ID DESC LIMIT 1"
	row := db.QueryRow(stmt)

	var user User
	if err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email); err != nil {
		handleError(err)
		return nil, err
	} else {
		return &user, nil
	}
}

func Delete(u *User) error {
	db := getDB()
	defer db.Close()

	stmt := "DELETE FROM USERS where id = ?"
	if _, err := db.Exec(stmt, u.ID); err != nil {
		handleError(err)
		return err
	} else {
		fmt.Println("Deletion completed")
		return nil
	}
}

func FindByID(n int) (*User, error) {
	db := getDB()
	defer db.Close()
	stmt := "SELECT ID, FIRST_NAME, LAST_NAME, EMAIL FROM USERS where id = ?"
	row := db.QueryRow(stmt, n)

	var user User
	if err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email); err != nil {
		handleError(err)
		return nil, err
	} else {
		return &user, nil
	}
}

func All() ([]User, error) {
	db := getDB()
	defer db.Close()

	var rowCount int
	row := db.QueryRow("SELECT COUNT(*) FROM USERS")
	if err := row.Scan(&rowCount); err != nil {
		return nil, err
	}

	users := make([]User, 0, rowCount)
	queryStmt := "SELECT ID, FIRST_NAME, LAST_NAME, EMAIL FROM USERS ORDER BY ID"
	if rows, err := db.Query(queryStmt); err != nil {
		handleError(err)
		return users, err
	} else {
		for rows.Next() {
			var user User
			rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email)
			users = append(users, user)
		}
		return users, nil
	}
}

func handleError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func getDB() *sql.DB {
	db, err := sql.Open("mysql", "root:fxrate@/GoDB")
	handleError(err)
	return db
}
*/
