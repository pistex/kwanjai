package controllers

import (
	"kwanjai/configuration"
	"kwanjai/helpers"
	"kwanjai/libraries"
	"kwanjai/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AllProject endpoint
func AllProject() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		username := helpers.GetUsername(ginContext)
		searchProjects, err := libraries.FirestoreSearch("projects", "Members", "array-contains", username)
		if err != nil {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		allProjects := []*models.Project{}
		for _, p := range searchProjects {
			project := new(models.Project)
			p.DataTo(project)
			project.ID = p.Ref.ID
			allProjects = append(allProjects, project)
		}
		ginContext.JSON(http.StatusOK, gin.H{"projects": allProjects})
		return
	}
}

// NewProject endpoint
func NewProject() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		username := helpers.GetUsername(ginContext)
		project := new(models.Project)
		err := ginContext.ShouldBindJSON(project)
		project.User = username
		if err != nil {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		// Check plan limit
		if helpers.ExceedProjectLimit(ginContext) {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Project exceed plan limit."})
			return
		}

		status, message, project := project.CreateProject()
		ginContext.JSON(status,
			gin.H{
				"message": message,
				"project": project,
			})
	}
}

// FindProject endpoint
func FindProject() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		project := new(models.Project)
		ginContext.ShouldBindJSON(project)
		if project.ID == "" {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID."})
			return
		}
		status, message, project := project.FindProject()
		ginContext.JSON(status,
			gin.H{
				"message": message,
				"project": project,
			})
	}
}

// UpdateProject endpoint
func UpdateProject() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		username := helpers.GetUsername(ginContext)
		project := new(models.Project)
		ginContext.ShouldBindJSON(project)
		_, memberIsIncludedOwner := libraries.Find(project.Members, username)
		if project.ID == "" || !memberIsIncludedOwner || project.Name == "" {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid UUID."})
			return
		}
		copiedProject := new(models.Project)
		copiedProject.ID = project.ID
		status, message, _ := copiedProject.FindProject()
		if status != http.StatusOK {
			ginContext.JSON(status,
				gin.H{
					"message": message,
				})
			return
		}
		if copiedProject.User != username {
			ginContext.JSON(http.StatusForbidden, gin.H{"message": "You cannot perform this action."})
			return
		}
		project.User = copiedProject.User
		status, message, project = project.UpdateProject()
		ginContext.JSON(status,
			gin.H{
				"message": message,
				"project": project,
			})
	}
}

// DeleteProject endpoint
func DeleteProject() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		username := helpers.GetUsername(ginContext)
		project := new(models.Project)
		ginContext.ShouldBindJSON(project)
		if project.ID == "" {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID."})
			return
		}
		copiedProject := new(models.Project)
		copiedProject.ID = project.ID
		status, message, _ := copiedProject.FindProject()
		if status != http.StatusOK {
			ginContext.JSON(status,
				gin.H{
					"message": message,
				})
			return
		}
		if copiedProject.User != username {
			ginContext.JSON(http.StatusForbidden, gin.H{"message": "You cannot perform this action."})
			return
		}
		status, message, _ = project.DeleteProject()
		db := libraries.FirestoreDB()
		searchBoard := db.Collection("boards").Where("Project", "==", project.ID).Documents(configuration.Context)
		allBoard, err := searchBoard.GetAll()
		if err != nil {
			log.Panic(err)
		}
		board := new(models.Board)
		for _, b := range allBoard {
			b.DataTo(board)
			board.ID = b.Ref.ID
			board.DeleteBoard()
		}
		db.Close()
		ginContext.JSON(status,
			gin.H{
				"message": message,
			})
	}
}
