// This project followed https://go.dev/doc/tutorial/web-service-gin
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type board struct {
	ID    int64  `json:"id"`
    Title string `json:"title"`
}

type card struct {
    ID      int64  `json:"id"`
    Content string `json:"content"`
    BoardID int64   `json:"bid"`
}

var db *sql.DB
var ctx context.Context

func getBoards(c *gin.Context) {
    var boards = []board {}
    boardRows, err := db.Query(`
        SELECT * FROM boards
    `)
    defer boardRows.Close()
    if err != nil {
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err})
        return
    }

    for boardRows.Next() {
        var newBoard board
        err := boardRows.Scan(&newBoard.ID, &newBoard.Title)
        if err != nil {
            c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err})
            return
        }
        boards = append(boards, newBoard)
    }

	c.IndentedJSON(http.StatusOK, boards)
}

func getBoardCards(c *gin.Context) {
    id := c.Param("id")
    var cards = []card {}

    cardRows, err := db.Query(`
        SELECT cards.id, content, bid
        FROM cards
        JOIN boards
        ON bid = boards.id
        WHERE bid = ?
    `, id)
    defer cardRows.Close()
    if err != nil {
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err})
        return
    }

    for cardRows.Next() {
        var newCard card
        err := cardRows.Scan(&newCard.ID, &newCard.Content, &newCard.BoardID)
        if err != nil {
            c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err})
            return
        }
        cards = append(cards, newCard)
    }

    c.IndentedJSON(http.StatusOK, cards)
}

func postBoard(c *gin.Context) {
    var newBoard board

    // Try to write json to object
    if err := c.BindJSON(&newBoard); err != nil {
        return
    }

    res, err := db.Exec("INSERT INTO boards (title) VALUES (?)", newBoard.Title)
    if (err != nil) {
        log.Fatal(err)
    }

    id, err := res.LastInsertId()
    newBoard.ID = id
    c.IndentedJSON(http.StatusCreated, newBoard)
}

func postCardToBoard(c *gin.Context) {
    id := c.Param("id") // get the id of board
    var newCard card

    if err := c.BindJSON(&newCard); err != nil {
        return
    }

    // verify foreign key is valid
    _, err := db.Exec("SELECT * FROM boards WHERE id = ?", id)
    if err != nil {
        c.IndentedJSON(http.StatusNotFound, gin.H{"message": "board not found"})
        return
    }

    res, err := db.Exec("INSERT INTO cards (content, bid) VALUES (?, ?)", newCard.Content, id)
    if (err != nil) {
        c.IndentedJSON(http.StatusNotFound, gin.H{"message": err})
        return
    }

    newCardID, err := res.LastInsertId()
    newCard.ID = newCardID
    newCard.BoardID, err = strconv.ParseInt(id, 10, 64)
    if err != nil {
        log.Fatal(err)
    }

    c.IndentedJSON(http.StatusCreated, newCard)
}

func getBoardByID(c *gin.Context) {
    id := c.Param("id")
    var newBoard board

    row := db.QueryRow("SELECT * FROM boards WHERE id = ?", id)
    err := row.Scan(&newBoard.ID, &newBoard.Title)

    switch err {
        case sql.ErrNoRows:
            c.IndentedJSON(http.StatusNotFound, gin.H{"message": "board not found"})
            return
        case nil:
            c.IndentedJSON(http.StatusOK, newBoard)
            return
    }

    c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err})
}

func deleteBoardByID(c *gin.Context) {
    id := c.Param("id")
    var newBoard board

    res, err := db.Exec("DELETE FROM boards WHERE id = ?", id)
    nrows, err := res.RowsAffected()

    switch nrows {
        case 0:
            c.IndentedJSON(http.StatusNotFound, gin.H{"message": "board not found"})
            return
        case 1:
            c.IndentedJSON(http.StatusOK, newBoard)
            return
    }

    c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err})
}

func createTables() {
    // setup card table
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS boards (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title VARCHAR(30) NOT NULL
        )
    `)
    if err != nil {
        log.Fatal(err)
    }

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS cards (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            content TEXT NOT NULL,
            bid INTEGER NOT NULL,
            FOREIGN KEY (bid) REFERENCES boards (id) ON DELETE CASCADE
        )
    `)
    if err != nil {
        log.Fatal(err)
    }
}


func main() {
    // setup database connection
    var err error
    db, err = sql.Open("sqlite3", "db.sqlite")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    createTables()
	router := gin.Default()
	router.GET("/", getBoards)
	router.POST("/", postBoard)
    router.GET("/:id/card", getBoardCards)
    router.POST("/:id/card", postCardToBoard)
    router.GET("/:id", getBoardByID)
    router.DELETE("/:id", deleteBoardByID)

	router.Run("localhost:8080")
}
