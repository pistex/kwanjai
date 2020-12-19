package models

import (
	"kwanjai/configuration"
	"kwanjai/libraries"
	"log"
	"net/http"
)

// Board model.
type Board struct {
	ID       string `json:"id"`
	User     string `json:"user"`
	Name     string `json:"name" binding:"required"`
	Project  string `json:"project" binding:"required"`
	Position int    `json:"position"`
}

// CreateBoard method returns status (int), message (string), board object.
func (board *Board) CreateBoard() (int, string, *Board) {
	reference, _, err := libraries.FirestoreAdd("boards", board)
	if err != nil {
		log.Panicln(err)
	}
	go libraries.FirestoreDeleteField("boards", reference.ID, "ID")
	board.ID = reference.ID
	return http.StatusCreated, "Created board.", board
}

// FindBoard method returns status (int), message (string), board object.
func (board *Board) FindBoard() (int, string, *Board) {
	if board.ID == "" {
		return http.StatusNotFound, "Board not found.", nil
	}
	getBoard, _ := libraries.FirestoreFind("boards", board.ID)
	if getBoard.Exists() {
		getBoard.DataTo(board)
		board.ID = getBoard.Ref.ID
		return http.StatusOK, "Get board successfully.", board
	}
	return http.StatusNotFound, "Board not found.", nil
}

// UpdateBoard board method.
func (board *Board) UpdateBoard(field string, value interface{}) (int, string, *Board) {
	_, err := libraries.FirestoreUpdateField("boards", board.ID, field, value)
	if err != nil {
		log.Panicln(err)
	}
	return http.StatusOK, "Updated sucessfully.", board
}

// DeleteBoard method returns status (int), message (string), nil object.
func (board *Board) DeleteBoard() (int, string, *Board) {
	_, err := libraries.FirestoreDelete("boards", board.ID)
	if err != nil {
		log.Panicln(err)
	}
	db := libraries.FirestoreDB()
	searchPost := db.Collection("posts").Where("Board", "==", board.ID).Documents(configuration.Context)
	allPost, err := searchPost.GetAll()
	if err != nil {
		log.Panic(err)
	}
	post := new(Post)
	for _, p := range allPost {
		p.DataTo(post)
		post.ID = p.Ref.ID
		post.DeletePost()
	}
	db.Close()
	return http.StatusOK, "Deleted sucessfully.", nil
}
