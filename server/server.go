package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type gridfsFile struct {
	Name   string `bson:"filename"`
	Length int64  `bson:"length"`
}

func uploadFileToMongoDB(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Подключение к базе данных MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:10000"))
	if err != nil {
		http.Error(w, "Ошибка подключения к базе данных", http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(context.Background())

	db := client.Database("db")
	opts := options.GridFSBucket().SetName("test-name")
	bucket, err := gridfs.NewBucket(db, opts)
	if err != nil {
		http.Error(w, "Ошибка создания bucket для GridFS", http.StatusInternalServerError)
		return
	}

	// Открытие потока для записи файла в базу данных MongoDB с использованием GridFS
	uploadStream, err := bucket.OpenUploadStream(header.Filename)
	if err != nil {
		http.Error(w, "Ошибка открытия потока для записи", http.StatusInternalServerError)
		return
	}
	defer uploadStream.Close()

	// Копирование содержимого файла из запроса в поток для записи в базу данных
	_, err = io.Copy(uploadStream, file)
	if err != nil {
		http.Error(w, "Ошибка при записи файла в базу данных", http.StatusInternalServerError)
		return
	}

	// Отправка ответа об успешной загрузке файла в базу данных
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Файл успешно сохранен в базе данных"))

	fmt.Printf("New file uploaded with ID %s\n", uploadStream.FileID)

	filter := bson.D{{"length", bson.D{{"$gt", 1}}}}
	cursor, _ := bucket.Find(filter)

	var foundFiles []gridfsFile
	if err = cursor.All(context.TODO(), &foundFiles); err != nil {
		http.Error(w, "Ошибка при записи файла в базу данных", http.StatusInternalServerError)
		return
	}
	for _, file := range foundFiles {
		fmt.Printf("filename: %s length: %d\n", file.Name, file.Length)
	}
}

func main() {

	http.HandleFunc("/upload", uploadFileToMongoDB)
	http.ListenAndServe(":8080", nil)
}
