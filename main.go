package main

import (
	"bank-account-api/api"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	_ "github.com/go-sql-driver/mysql"
)

type Server struct {
	DB *sql.DB
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
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		s.handleParamError(c, err)
	} else {
		var account api.BankAccount
		if err := c.ShouldBindJSON(&account); err != nil {
			s.handleParamError(c, err)
			return
		}
		account.UserID = id
		if err := s.accountService.Create(&account); err != nil {
			s.handleDBError(c, err)
			return
		} else {
			c.JSON(http.StatusCreated, account)
		}
	}
}

func (s *Server) DeleteBankAccount(c *gin.Context) {
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		s.handleParamError(c, err)
	} else {
		if err := s.accountService.Delete(id); err != nil {
			s.handleDBError(c, err)
		}
	}
}

func (s *Server) WithdrawBankAccount(c *gin.Context) {
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		s.handleParamError(c, err)
	} else {
		h := map[string]string{}
		if err := c.ShouldBindJSON(&h); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		amount, err := strconv.Atoi(h["amount"])
		if err != nil {
			s.handleParamError(c, err)
			return
		}
		if acc, err := s.accountService.Withdraw(id, amount); err != nil {
			s.handleDBError(c, err)
			return
		} else {
			c.JSON(http.StatusOK, acc)
		}
	}
}

func (s *Server) DepositBankAccount(c *gin.Context) {
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		s.handleParamError(c, err)
	} else {
		h := map[string]string{}
		if err := c.ShouldBindJSON(&h); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		amount, err := strconv.Atoi(h["amount"])
		if err != nil {
			s.handleParamError(c, err)
			return
		}
		if acc, err := s.accountService.Deposit(id, amount); err != nil {
			s.handleDBError(c, err)
			return
		} else {
			c.JSON(http.StatusOK, acc)
		}
	}
}

func (s *Server) Transfer(c *gin.Context) {
	h := map[string]string{}
	if err := c.ShouldBindJSON(&h); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	amount, err := strconv.Atoi(h["amount"])
	if err != nil {
		s.handleParamError(c, err)
		return
	}
	fromAcc, err := strconv.Atoi(h["from"])
	if err != nil {
		s.handleParamError(c, err)
		return
	}
	toAcc, err := strconv.Atoi(h["to"])
	if err != nil {
		s.handleParamError(c, err)
		return
	}
	err = s.accountService.Transfer(amount, fromAcc, toAcc)
	if err != nil {
		s.handleDBError(c, err)
	}
}

func (s *Server) Authenticate(c *gin.Context) {
	log.Printf("Request %v %v\n", c.Request.Method, c.Request.URL)
	apiKey := c.Request.Header.Get("key")
	row := s.DB.QueryRow("SELECT `key` FROM `KEY` WHERE `key` = ?", apiKey)
	if err := row.Scan(&apiKey); err == nil {
		return
	}
	c.AbortWithStatus(http.StatusUnauthorized)
}

func setupRoute(s *Server) *gin.Engine {
	r := gin.Default()
	
	r.POST("/transfers", s.Transfer)
	
	userGroup := r.Group("/users")
	userGroup.Use(s.Authenticate)
	userGroup.GET("", s.GetAllUsers)
	userGroup.GET("/:id", s.GetUserByID)
	userGroup.POST("", s.CreateNewUser)
	userGroup.PUT("/:id", s.UpdateUser)
	userGroup.DELETE("/:id", s.DeleteUserByID)
	userGroup.GET("/:id/bankAccounts", s.GetBankAccountsByUserID)
	userGroup.POST("/:id/bankAccounts", s.CreateNewBankAccount)

	accountGroup := r.Group("/bankAccounts")
	accountGroup.Use(s.Authenticate)
	accountGroup.DELETE("/:id", s.DeleteBankAccount)
	accountGroup.PUT("/:id/withdraw", s.WithdrawBankAccount)
	accountGroup.PUT("/:id/deposit", s.DepositBankAccount)
	
	// Test Command Example: 
	// curl -XPOST https://localhost:8080/admin/secrets -u admin:1234 -d "{\"key\": \"foobar\"}""
	adminGroup := r.Group("/admin")
	adminGroup.Use(gin.BasicAuth(gin.Accounts{
		"admin": "1234",
	}))
	adminGroup.POST("/secrets", s.CreateSecretKey)
	
	return r
}

func main() {
	// Testing Local Database URL: "root:fxrate@/GoDB"
	db, err := sql.Open("mysql", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	s := &Server{
		DB: db,
		userService: &api.UserServiceMySQL{
			DB: db,
		},
		accountService: &api.AccountServiceMySQL{
			DB: db,
		},
	}
	r := setupRoute(s)
	r.Run(":" + os.Getenv("PORT"))
}

type Secret struct {
	ID  int64
	Key string
}


func (s *Server) CreateSecretKey(c *gin.Context) {
	var secret Secret
	err := c.ShouldBindJSON(&secret)
	if err != nil {
		s.handleParamError(c, err)
		return
	}
	res, err := s.DB.Exec("INSERT INTO `KEY` (`key`) values (?)", secret.Key)
	if err != nil {
		s.handleDBError(c, err)
		return
	}
	i, err := res.LastInsertId()
	if err != nil {
		s.handleDBError(c, err)
		return
	}
	secret.ID = i
	c.JSON(http.StatusCreated, secret)
}



