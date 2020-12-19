package controllers

import (
	"kwanjai/configuration"
	"kwanjai/helpers"
	"kwanjai/libraries"
	"kwanjai/models"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

// AllBoard endpoint
func AllBoard() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		username := helpers.GetUsername(ginContext)
		board := new(models.Board)
		ginContext.ShouldBindJSON(board)
		if board.Project == "" {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Project UUID is reqired."})
			return
		}

		searchBoards, err := libraries.FirestoreSearch("boards", "Project", "==", board.Project)
		if err != nil {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		allBoards := []*models.Board{}
		for _, b := range searchBoards {
			board := new(models.Board)
			b.DataTo(board)
			// check project membership
			if !helpers.IsProjectMember(username, board.Project) {
				ginContext.JSON(http.StatusForbidden, gin.H{"message": "You cannot perform this action."})
				return
			}
			board.ID = b.Ref.ID
			allBoards = append(allBoards, board)
		}
		ginContext.JSON(http.StatusOK,
			gin.H{
				"boards": allBoards,
			})
	}
}

// NewBoard endpoint
func NewBoard() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		username := helpers.GetUsername(ginContext)
		board := new(models.Board)
		err := ginContext.ShouldBindJSON(board)
		board.User = username
		if err != nil {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Check project ownership
		if !helpers.IsOwner(username, "projects", board.Project) {
			ginContext.JSON(http.StatusForbidden, gin.H{"message": "You cannot perform this action."})
			return
		}

		// Get all board
		allBoards, err := libraries.FirestoreSearch("boards", "Project", "==", board.Project)
		if err != nil {
			log.Panic(nil)
		}
		boardNumber := len(allBoards)

		// Check plan limit
		if helpers.ExceedBoardLimit(ginContext, boardNumber) {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Board exceed plan limit."})
			return
		}

		board.Position = boardNumber + 1
		status, message, _ := board.CreateBoard()
		ginContext.JSON(status,
			gin.H{
				"message": message,
				"board":   board,
			})
	}
}

// FindBoard endpoint
func FindBoard() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		username := helpers.GetUsername(ginContext)
		board := new(models.Board)
		ginContext.ShouldBindJSON(board)
		if board.ID == "" {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID."})
			return
		}
		status, message, _ := board.FindBoard()

		// Check project membership
		if !helpers.IsProjectMember(username, board.Project) {
			ginContext.JSON(http.StatusForbidden, gin.H{"message": "You cannot perform this action."})
			return
		}

		ginContext.JSON(status,
			gin.H{
				"message": message,
				"board":   board,
			})
	}
}

// UpdateBoard endpoint
func UpdateBoard() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		username := helpers.GetUsername(ginContext)
		updatedBoard := new(models.Board)
		ginContext.ShouldBindJSON(updatedBoard)
		if updatedBoard.ID == "" || updatedBoard.Name == "" {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid board update form."})
			return
		}

		// Get old board
		oldBoard := new(models.Board)
		oldBoard.ID = updatedBoard.ID
		status, message, _ := oldBoard.FindBoard()
		if status != http.StatusOK {
			ginContext.JSON(status, gin.H{"message": message})
			return
		}

		// Check board ownership
		if username != oldBoard.User {
			ginContext.JSON(http.StatusForbidden, gin.H{"message": "You cannot perform this action."})
			return
		}

		// Check if position is changed
		oldPosition := oldBoard.Position
		newPosition := updatedBoard.Position
		// If board position is changed
		if oldPosition != newPosition {
			if newPosition != oldPosition+1 && newPosition != oldPosition-1 {
				ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Forbidden board position."})
				return
			}
			projectID := oldBoard.Project
			db := libraries.FirestoreDB()
			// #1 find the board where board.Position = newPosition and change board position to old position
			searchBoardWithNewPostion := db.Collection("boards").Where("Project", "==", projectID).Where("Position", "==", newPosition).Documents(configuration.Context)
			boardWithNewPostion, err := searchBoardWithNewPostion.GetAll()
			if err != nil {
				log.Panic(err)
			}
			_, err = db.Collection("boards").Doc(boardWithNewPostion[0].Ref.ID).Update(configuration.Context, []firestore.Update{
				{
					Path:  "Position",
					Value: oldPosition,
				},
			})
			if err != nil {
				log.Panic(err)
			}
			// #2 change current board position to new position.
			_, err = db.Collection("boards").Doc(updatedBoard.ID).Update(configuration.Context, []firestore.Update{
				{
					Path:  "Position",
					Value: newPosition,
				},
			})
			if err != nil {
				log.Panic(err)
			}
			// example
			// case: move 3 to 2 of 4
			// #0 1 2 3 4
			// #1 1 3 3 4
			// #2 1 3 2 4
			// case: move 3 to 4 of 4
			// #0 1 2 3 4
			// #1 1 2 3 3
			// #2 1 2 4 3
			db.Close()
		}
		if oldBoard.Name != updatedBoard.Name {
			status, message, _ = updatedBoard.UpdateBoard("Name", updatedBoard.Name)
			if status != http.StatusOK {
				ginContext.JSON(status, gin.H{"message": message})
				return
			}
		}
		status, message, _ = updatedBoard.FindBoard()
		ginContext.JSON(status,
			gin.H{
				"message": message,
				"board":   updatedBoard,
			})
	}
}

// DeleteBoard endpoint
func DeleteBoard() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		username := helpers.GetUsername(ginContext)
		board := new(models.Board)
		ginContext.ShouldBindJSON(board)
		status, message, _ := board.FindBoard()
		if status != http.StatusOK {
			ginContext.JSON(status, gin.H{"message": message})
			return
		}

		// Check board ownership
		if !helpers.IsOwner(username, "boards", board.ID) {
			ginContext.JSON(http.StatusForbidden, gin.H{"message": "You cannot perform this action."})
			return
		}
		status, message, _ = board.DeleteBoard()
		if status != http.StatusOK {
			ginContext.JSON(status, gin.H{"message": message})
			return
		}

		deletedPosition := board.Position
		projectID := board.Project
		allBoards, err := libraries.FirestoreSearch("boards", "Project", "==", projectID)
		if err != nil {
			log.Panicln(err)
		}
		board = new(models.Board)
		for _, b := range allBoards {
			b.DataTo(board)
			if board.Position > deletedPosition {
				libraries.FirestoreUpdateField("boards", b.Ref.ID, "Position", board.Position-1)
			}
		}
		ginContext.JSON(status,
			gin.H{
				"message": message,
			})
	}
}
