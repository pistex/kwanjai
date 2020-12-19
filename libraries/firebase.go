package libraries

import (
	"errors"
	"image"
	"image/png"
	"kwanjai/configuration"
	"log"
	"mime/multipart"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/nfnt/resize"
	"google.golang.org/api/option"
)

// FirebaseApp initialize firebase by credential.json.
func FirebaseApp() *firebase.App {
	var err error
	var app *firebase.App
	if os.Getenv("GIN_MODE") == "release" {
		conf := &firebase.Config{
			ProjectID:     configuration.FirebaseProjectID,
			StorageBucket: "kwanjai-a3803.appspot.com",
		}
		app, err = firebase.NewApp(configuration.Context, conf)
	} else {
		conf := &firebase.Config{
			StorageBucket: "kwanjai-a3803.appspot.com",
		}
		sa := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
		app, err = firebase.NewApp(configuration.Context, conf, sa)
	}
	if err != nil {
		log.Panicln(err)
	}
	return app
}

// FirestoreDB return firestore client and error
func FirestoreDB() *firestore.Client {
	firestoreClient, err := FirebaseApp().Firestore(configuration.Context)
	if err != nil {
		log.Panic(err)
	}
	return firestoreClient
}

// FirestoreFind by collection and document ID.
func FirestoreFind(collecttion string, id string) (*firestore.DocumentSnapshot, error) {
	if collecttion == "" || id == "" {
		// create blank document
		blank := new(firestore.DocumentSnapshot)
		blank.Ref = new(firestore.DocumentRef)
		blank.Ref.Parent = new(firestore.CollectionRef)
		return blank, errors.New("invalid document reference")
	}
	firestoreClient, err := FirebaseApp().Firestore(configuration.Context)
	defer firestoreClient.Close()
	document, err := firestoreClient.Collection(collecttion).Doc(id).Get(configuration.Context)
	return document, err
}

// FirestoreDelete by collection and document ID.
func FirestoreDelete(collecttion string, id string) (*firestore.WriteResult, error) {
	if collecttion == "" || id == "" {
		// create blank result
		blank := new(firestore.WriteResult)
		return blank, errors.New("invalid document reference")
	}
	firestoreClient, err := FirebaseApp().Firestore(configuration.Context)
	defer firestoreClient.Close()
	result, err := firestoreClient.Collection(collecttion).Doc(id).Delete(configuration.Context)
	return result, err
}

// FirestoreSearch by collection and condition
func FirestoreSearch(collecttion string, field string, condition string, property interface{}) ([]*firestore.DocumentSnapshot, error) {
	firestoreClient, err := FirebaseApp().Firestore(configuration.Context)
	defer firestoreClient.Close()
	search := firestoreClient.Collection(collecttion).Where(field, condition, property).Documents(configuration.Context)
	documents, err := search.GetAll()
	return documents, err
}

// FirestoreCreateOrSet by collection, id.
func FirestoreCreateOrSet(collecttion string, id string, data interface{}) (*firestore.WriteResult, error) {
	if collecttion == "" || id == "" {
		// create blank result
		blank := new(firestore.WriteResult)
		return blank, errors.New("invalid document reference")
	}
	firestoreClient, err := FirebaseApp().Firestore(configuration.Context)
	defer firestoreClient.Close()
	result, err := firestoreClient.Collection(collecttion).Doc(id).Set(configuration.Context, data)
	return result, err
}

// FirestoreAdd by collection and automatically create id.
func FirestoreAdd(collecttion string, data interface{}) (*firestore.DocumentRef, *firestore.WriteResult, error) {
	if collecttion == "" {
		// create blank result
		blankResult := new(firestore.WriteResult)
		blankReference := new(firestore.DocumentRef)
		return blankReference, blankResult, errors.New("invalid document reference")
	}
	firestoreClient, err := FirebaseApp().Firestore(configuration.Context)
	defer firestoreClient.Close()
	reference, result, err := firestoreClient.Collection(collecttion).Add(configuration.Context, data)
	return reference, result, err
}

// FirestoreUpdateField by collection, id, and field.
func FirestoreUpdateField(collecttion string, id string, field string, property interface{}) (*firestore.WriteResult, error) {
	if collecttion == "" || id == "" || field == "" {
		// create blank result
		blank := new(firestore.WriteResult)
		return blank, errors.New("invalid document reference")
	}
	firestoreClient, err := FirebaseApp().Firestore(configuration.Context)
	defer firestoreClient.Close()
	result, err := firestoreClient.Collection(collecttion).Doc(id).Update(configuration.Context, []firestore.Update{
		{
			Path:  field,
			Value: property,
		},
	})
	return result, err
}

// FirestoreUpdateFieldIfNotBlank by collection, id, and field.
func FirestoreUpdateFieldIfNotBlank(collecttion string, id string, field string, property interface{}) (*firestore.WriteResult, error) {
	if collecttion == "" || id == "" || field == "" {
		// create blank result
		blank := new(firestore.WriteResult)
		return blank, errors.New("invalid document reference")
	}
	if property.(string) == "" {
		blank := new(firestore.WriteResult)
		return blank, nil
	}
	firestoreClient, err := FirebaseApp().Firestore(configuration.Context)
	defer firestoreClient.Close()
	result, err := firestoreClient.Collection(collecttion).Doc(id).Update(configuration.Context, []firestore.Update{
		{
			Path:  field,
			Value: property,
		},
	})
	return result, err
}

// FirestoreDeleteField by collection, id, and field.
func FirestoreDeleteField(collecttion string, id string, field string) (*firestore.WriteResult, error) {
	firestoreClient, err := FirebaseApp().Firestore(configuration.Context)
	defer firestoreClient.Close()
	result, err := firestoreClient.Collection(collecttion).Doc(id).Update(configuration.Context, []firestore.Update{
		{
			Path:  field,
			Value: firestore.Delete,
		},
	})
	return result, err
}

// CloudStorageUpload to
func CloudStorageUpload(file multipart.File, path string) {
	image, _, err := image.Decode(file)
	if err != nil {
		log.Panicln(err)
	}
	file.Close()
	storageClient, err := FirebaseApp().Storage(configuration.Context)
	if err != nil {
		log.Panicln(err)
	}
	newImage := resize.Thumbnail(200, 200, image, resize.Lanczos3)
	bucket, err := storageClient.DefaultBucket()
	if err != nil {
		log.Panicln(err)
	}
	fileWriter := bucket.Object(path).NewWriter(configuration.Context)
	png.Encode(fileWriter, newImage)
	fileWriter.Close()
}

func CreateProfilePicture(username string) {
	storageClient, err := FirebaseApp().Storage(configuration.Context)
	if err != nil {
		log.Panicln(err)
	}
	bucket, err := storageClient.DefaultBucket()
	if err != nil {
		log.Panicln(err)
	}
	src := bucket.Object("anonymous.png")
	dst := bucket.Object(username + ".png")
	_, err = dst.CopierFrom(src).Run(configuration.Context)
	if err != nil {
		log.Panicln(err)
	}
}
