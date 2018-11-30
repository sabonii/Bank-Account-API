package main

import (
	"bank-account-api/api"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Server struct {
	userService api.UserService
	accountService api.AccountService
}

func (server *Server) handleDBError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"object":  "error",
		"message": fmt.Sprintf("db: query error: %s", err),
	})
}

func (server *Server) handleParamError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"object":  "error",
		"message": fmt.Sprintf("json: wrong params: %s", err),
	})
}

func (s *Server) GetAllUsers(c *gin.Context) {
	if users, err := s.userService.All(); err != nil {
		s.handleDBError(c, err)
		return
	} else {
		c.JSON(http.StatusOK, users)
	}
}

func (s *Server) GetUserByID(c *gin.Context) {
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		s.handleParamError(c, err)
	} else {
		if user, err := s.userService.FindByID(id); err != nil {
			s.handleDBError(c, err)
			return
		} else {
			c.JSON(http.StatusOK, user)
		}
	}
}

func (s *Server) DeleteUserByID(c *gin.Context) {
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		s.handleDBError(c, err)
	} else {
		if err := s.userService.Delete(id); err != nil {
			s.handleDBError(c, err)
		}
	}
}

func (s *Server) UpdateUser(c *gin.Context) {
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		s.handleDBError(c, err)
	} else {
		var user api.User
		if err := c.ShouldBindJSON(&user); err != nil {
			s.handleParamError(c, err)
			return
		}
		user.ID = id
		if err = s.userService.Update(&user); err != nil {
			s.handleDBError(c, err)
			return
		}
	}
}

func (s *Server) CreateNewUser(c *gin.Context) {
	var user api.User
	if err := c.ShouldBindJSON(&user); err != nil {
		s.handleParamError(c, err)
		return
	}
	if err := s.userService.Insert(&user); err != nil {
		s.handleDBError(c, err)
		return
	} else {
		c.JSON(http.StatusCreated, user)
	}
}

func (s *Server) Transfer(c *gin.Context) {
}

func (s *Server) GetBankAccountsByUserID(c *gin.Context) {
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		s.handleParamError(c, err)
	} else {
		if accounts, err := s.accountService.List(id); err != nil {
			s.handleDBError(c, err)
			return
		} else {
			c.JSON(http.StatusOK, accounts)
		}
	}
}

func (s *Server) CreateNewBankAccount(c *gin.Context) {
}

func (s *Server) DeleteBankAccount(c *gin.Context) {
}

func (s *Server) WithdrawBankAccount(c *gin.Context) {
}

func (s *Server) DepositBankAccount(c *gin.Context) {
}

/*func (s *Server) AuthToDo(c *gin.Context) {
	if user, _, ok := c.Request.BasicAuth(); ok {
		var id int64
		row := s.db.QueryRow("SELECT count(*) FROM secrets WHERE key = $1", user)
		if err := row.Scan(&id); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"object":  "error",
				"message": fmt.Sprintf("db: query error: %s", err),
			})
		} else if id != 1 {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}*/

func setupRoute(s *Server) *gin.Engine {
	r := gin.Default()

	r.POST("/transfers", s.Transfer)
	
	userGroup := r.Group("/users")
	/*
	userGroup.Use(s.AuthToDo)
	*/
	userGroup.GET("", s.GetAllUsers)
	userGroup.GET("/:id", s.GetUserByID)
	userGroup.POST("", s.CreateNewUser)
	userGroup.PUT("/:id", s.UpdateUser)
	userGroup.DELETE("/:id", s.DeleteUserByID)
	userGroup.GET("/:id/bankAccounts", s.GetBankAccountsByUserID)
	userGroup.POST("/:id/bankAccounts", s.CreateNewBankAccount)

	accountGroup := r.Group("/bankAccounts")
	accountGroup.DELETE("/:id", s.DeleteBankAccount)
	accountGroup.PUT("/:id/withdraw", s.WithdrawBankAccount)
	accountGroup.PUT("/:id/deposit", s.DepositBankAccount)
	
	// curl -XPOST https://localhost:8080/admin/secrets -u admin:1234 -d "{\"key\": \"foobar\"}""
	adminGroup := r.Group("/admin")
	adminGroup.Use(gin.BasicAuth(gin.Accounts{
		"admin": "1234",
	}))
	// adminGroup.POST("/secrets", s.CreateSecret)
	
	return r
}

func main() {
	db, err := sql.Open("mysql", "root:fxrate@/GoDB")
	if err != nil {
		log.Fatal(err)
	}

	/*
		createTable := `
		CREATE TABLE IF NOT EXISTS todos (
			id SERIAL PRIMARY KEY,
			todo TEXT,
			created_at TIMESTAMP WITHOUT TIME ZONE,
			updated_at TIMESTAMP WITHOUT TIME ZONE
		);
		CREATE TABLE IF NOT EXISTS secrets (
			id SERIAL PRIMARY KEY,
			key TEXT
		);
		`
		if _, err := db.Exec(createTable); err != nil {
			log.Fatal(err)
		}
	*/

	s := &Server{
		userService: &api.UserServiceMySQL{
			DB: db,
		},
		accountService: &api.AccountServiceMySQL{
			DB: db,
		},
	}
	r := setupRoute(s)
	r.Run(":8080")
}

type Secret struct {
	ID  int64
	Key string
}

/*
func (s *Server) CreateSecret(c *gin.Context) {
	var secret Secret
	err := c.ShouldBindJSON(&secret)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("json: wrong params: %s", err),
		})
		return
	}
	row := s.db.QueryRow("INSERT INTO `KEY` (key) values ($1) RETURNING id", secret.Key)

	if err := row.Scan(&secret.ID); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("db: query error: %s", err),
		})
		return
	}

	c.JSON(http.StatusCreated, secret)
}
*/
