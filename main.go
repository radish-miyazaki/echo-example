package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func initDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func validateUser(name string, age int) error {
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is empty")
	}

	if len(name) >= 100 {
		return echo.NewHTTPError(http.StatusBadRequest, "name is too long")
	}

	if age < 0 || age >= 200 {
		return echo.NewHTTPError(http.StatusBadRequest, "age must be between 0 and 200")
	}

	return nil
}

func main() {
	db := initDB("example.db")
	e := echo.New()
	e.Use(middleware.Logger())

	e.POST("/users", func(c echo.Context) error {
		name := c.FormValue("name")
		age, err := strconv.Atoi(c.FormValue("age"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		if err = validateUser(name, age); err != nil {
			return err
		}

		result, err := db.Exec("INSERT INTO users (name, age) VALUES (?, ?)", name, age)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		id, _ := result.LastInsertId()

		return c.JSON(http.StatusOK, &User{ID: int(id), Name: name, Age: age})
	})

	e.PUT("/users/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		name := c.FormValue("name")
		age, err := strconv.Atoi(c.FormValue("age"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		if err = validateUser(name, age); err != nil {
			return err
		}

		result, err := db.Exec("UPDATE users SET name = ?, age = ? WHERE id = ?", name, age, id)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}

		return c.JSON(http.StatusOK, &User{ID: id, Name: name, Age: age})
	})

	e.GET("/users", func(c echo.Context) error {
		rows, err := db.Query("SELECT id, name, age FROM users")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		defer rows.Close()

		users := []User{}
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.ID, &user.Name, &user.Age); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			users = append(users, user)
		}

		return c.JSON(http.StatusOK, users)
	})

	e.DELETE("/users/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		result, err := db.Exec("DELETE FROM users WHERE id = ?", id)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}

		return c.NoContent(http.StatusNoContent)
	})

	e.GET("/users/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		row := db.QueryRow("SELECT id, name, age FROM users WHERE id = ?", id)

		var user User
		if err := row.Scan(&user.ID, &user.Name, &user.Age); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, user)
	})

	e.Start(":8080")
}
