package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	_ "image/jpeg"
	_ "image/png"
	"kwanjai/configuration"
	"kwanjai/controllers"
	"kwanjai/libraries"
	"kwanjai/middlewares"
	"kwanjai/models"
	"log"
	"os"
	"time"
)

//setupServer
func setupServer() {
	var err error
	if os.Getenv("GIN_MODE") == "" {
		if err = os.Setenv("GIN_MODE", "debug"); err != nil {
			log.Fatalln(err.Error())
		}
	}
	if configuration.BaseDirectory, err = os.Getwd(); err != nil {
		log.Fatalln(err.Error())
	}

	if os.Getenv("PORT") != "" {
		configuration.Port = ":" + os.Getenv("PORT")
	} else {
		configuration.Port = ":8080"
	}

	libraries.InitializeGCP() // BaseDirectory need to be set before initialization.
	configuration.Context = context.Background()
	configuration.FrontendURL = "https://kwanjai.pistex.dev"
	configuration.BackendURL = "https://kwanjai.pistex.dev/api"
	configuration.FirebaseProjectID = "kwanjai-a3803"
	// Authentication settings
	configuration.DefaultAuthenticationBackend = middlewares.JWTAuthorization()
	// JWT settings
	//if configuration.JWTAccessTokenSecretKey, err = libraries.AccessSecretVersion(
	//	"projects/978676563951/secrets/JWTAccessTokenSecretKey/versions/1"); err != nil {
	//	log.Fatalln(err.Error())
	//}
	//if configuration.JWTRefreshTokenSecretKey, err = libraries.AccessSecretVersion(
	//	"projects/978676563951/secrets/JWTRefreshTokenSecretKey/versions/1"); err != nil {
	//	log.Fatalln(err.Error())
	//}
	configuration.JWTAccessTokenSecretKey = "access"
	configuration.JWTAccessTokenSecretKey = "refresh"
	configuration.JWTAccessTokenLifetime = time.Hour * 2
	configuration.JWTRefreshTokenLifetime = time.Hour * 8

	//// Email service settings
	//if configuration.EmailServicePassword, err = libraries.AccessSecretVersion(
	//	"projects/978676563951/secrets/EmailServicePassword/versions/1"); err != nil {
	//	log.Fatalln(err.Error())
	//}
	//configuration.EmailVerificationLifetime = time.Hour * 24 * 7
	//
	//// Payment gateway settings
	//if configuration.OmisePublicKey, err = libraries.AccessSecretVersion(
	//	"projects/978676563951/secrets/OmisePublicKey/versions/1"); err != nil {
	//	log.Fatalln(err.Error())
	//}
	//if configuration.OmiseSecretKey, err = libraries.AccessSecretVersion(
	//	"projects/978676563951/secrets/OmiseSecretKey/versions/1"); err != nil {
	//	log.Fatalln(err.Error())
	//}


	// Database setup
	configuration.SQLUsername = "username"
	configuration.SQLPassword = "password"
	dbHost := os.Getenv("DATABASE_HOST")
	configuration.SQLHostname = fmt.Sprintf("tcp(%v:3306)", dbHost)
	log.Printf("Connecting to %v",configuration.SQLHostname)
	configuration.SQLDatabaseName = "database"
	if configuration.SQL, err = gorm.Open(
		mysql.Open(fmt.Sprintf("%v:%v@%v/%v",
			configuration.SQLUsername,
			configuration.SQLPassword,
			configuration.SQLHostname,
			configuration.SQLDatabaseName)),
		nil); err != nil{
		log.Fatalln(err.Error())
	}
	if err = configuration.SQL.AutoMigrate(&models.User{}); err !=nil{
		log.Fatalln(err.Error())
	}
}

func getServer(mode string) *gin.Engine {
	if mode == "debug" {
		log.Println("running in debug mode.")
	} else if mode == "test" {
		gin.SetMode(gin.TestMode)
		log.Println("running in test mode.")
	} else if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
		log.Println("running in production mode.")
	}
	ginEngine := gin.Default()
	ginEngine.Use(configuration.DefaultAuthenticationBackend)
	authentication := ginEngine.Group("/authentication")
	authentication.POST("/register", controllers.RegisterUser())
	//authentication.POST("/login", controllers.Login())
	//authentication.POST("/register", controllers.Register())
	//authentication.POST("/logout", middlewares.AuthenticatedOnly(), controllers.Logout())
	//authentication.POST("/verify_email/:ID", controllers.VerifyEmail())
	//authentication.POST("/resend_verification_email", controllers.ResendVerifyEmail())
	//authentication.POST("/token/refresh", controllers.RefreshToken())
	//authentication.GET("/token/verify", middlewares.AuthenticatedOnly(), controllers.TokenVerification())
	//user := ginEngine.Group("/user")
	//user.Use(middlewares.AuthenticatedOnly())
	//user.GET("/all", controllers.AllUsernames())
	//user.GET("/my_profile", controllers.MyProfile())
	//user.POST("/update_password", controllers.PasswordUpdate())
	//user.PATCH("/profile_picture", controllers.ProfilePicture())
	//user.PATCH("/update_profile", controllers.UpdateProfile())
	//user.POST("/pay", controllers.UpgradePlan())
	//user.POST("/unsubscribe", controllers.Unsubscribe())
	//project := ginEngine.Group("/project")
	//project.Use(middlewares.AuthenticatedOnly())
	//{
	//	project.GET("/all", controllers.AllProject())
	//	project.POST("/new", controllers.NewProject())
	//	project.POST("/find", controllers.FindProject())
	//	project.PATCH("/update", controllers.UpdateProject())
	//	project.DELETE("/delete", controllers.DeleteProject())
	//}
	//board := ginEngine.Group("/board")
	//board.Use(middlewares.AuthenticatedOnly())
	//{
	//	board.POST("/all", controllers.AllBoard())
	//	board.POST("/new", controllers.NewBoard())
	//	board.POST("/find", controllers.FindBoard())
	//	board.PATCH("/update", controllers.UpdateBoard())
	//	board.DELETE("/delete", controllers.DeleteBoard())
	//}
	//post := ginEngine.Group("/post")
	//post.Use(middlewares.AuthenticatedOnly())
	//{
	//	post.POST("/all", controllers.AllPost())
	//	post.POST("/new", controllers.NewPost())
	//	post.PATCH("/update", controllers.UpdatePost())
	//	post.DELETE("/delete", controllers.DeletePost())
	//	post.POST("/comment/new", controllers.NewComment())
	//	post.PATCH("/comment/update", controllers.UpdateComment())
	//	post.DELETE("/comment/delete", controllers.DeleteComment())
	//}
	return ginEngine
}

func main() {
	setupServer()
	ginEngine := getServer(os.Getenv("GIN_MODE"))
	if err := ginEngine.Run(configuration.Port); err != nil {
		log.Fatalln(err.Error())
	}
}
