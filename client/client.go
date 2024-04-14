package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type Args struct {
	Dbn   string
	Buckn string
	File  *os.File
	Fn    string
}

func sendFileToServer(filepath string, targetURL string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Создание тела запроса с файлом
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	writer.Close()

	// Создание POST-запроса на сервер
	req, err := http.NewRequest("POST", targetURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Отправка запроса на сервер
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("неправильный статус ответа: %v", resp.Status)
	}

	return nil
}

func main() {
	filepath := "30973089.pdf"
	targetURL := "http://localhost:8080/upload"

	err := sendFileToServer(filepath, targetURL)
	if err != nil {
		fmt.Println("Ошибка отправки файла:", err)
	} else {
		fmt.Println("Файл успешно отправлен на сервер")
	}
}
