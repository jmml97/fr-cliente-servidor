package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func handleError(err error, text string, code int) {

	if err != nil {
		fmt.Println(text)
		fmt.Println("err:", err)
		os.Exit(code)
	}

}

func receiveImage(conn net.Conn) bytes.Buffer {

	var receivedImage bytes.Buffer

	// Recibimos primero el tamaño de la imagen
	var receivedImageSize int64
	err := binary.Read(conn, binary.LittleEndian, &receivedImageSize)
	handleError(err, "Error durante la recepción del tamaño de la imagen", 1)

	fmt.Println("Tamaño de la imagen recibida:", receivedImageSize)

	// Ahora recibimos la imagen en sí
	n, err := io.CopyN(&receivedImage, conn, receivedImageSize)
	handleError(err, "Error durante la recepción de la imagen", 1)

	fmt.Println(n, "bytes recibidos")
	return receivedImage

}

func sendImage(conn net.Conn, imageFile *os.File) {

	imageInfo, _ := imageFile.Stat()
	imageSize := int64(imageInfo.Size())

	fmt.Println("Tamaño de la imagen a enviar:", imageSize)

	//sender
	err := binary.Write(conn, binary.LittleEndian, imageSize)
	handleError(err, "Error al enviar el tamaño de la imagen al servidor", 2)

	n, err := io.CopyN(conn, imageFile, imageSize)
	handleError(err, "Error durante la recepción de la imagen", 2)

	fmt.Println(n, "bytes enviados")

}

func reveiveMenu(conn net.Conn) {

	// \f simboliza fin del mensaje
	menuMessage, _ := bufio.NewReader(conn).ReadString('\f')
	fmt.Print("Mensaje del servidor:\n" + menuMessage)

}

func writeOption(conn net.Conn) int {

	// Creamos un lector para leer desde stdin
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	option := scanner.Text()

	conn.Write([]byte(option + "#"))

	if option == "3" {
		fmt.Println("Ejecución cancelada")
		os.Exit(10)
	}

	optionInt, _ := strconv.Atoi(option)

	return optionInt
}

func main() {

	// Nos conectamos al socket tcp
	conn, err := net.Dial("tcp", "127.0.0.1:8081")
	handleError(err, "No se ha podido establecer la conexión con el servidor", 3)
	defer conn.Close()

	// Bucle del programa. Podemos convertir imágenes hasta que decidamos
	// terminar la ejecución del programa
	for {

		// Mostramos el menú que recibe del servidor
		reveiveMenu(conn)

		// Pedimos al usuario la opción que se enviará al servidor
		option := writeOption(conn)

		var outExtension string

		switch option {
		case 1:
			outExtension = ".jpeg"
		case 2:
			outExtension = ".png"
		}

		// Creamos un lector para leer desde stdin
		scanner := bufio.NewScanner(os.Stdin)

		// Leemos el nombre del archivo
		fmt.Println("Introduce la ruta del archivo a convertir")
		scanner.Scan()
		filename := scanner.Text()

		// Leemos el archivo de la imagen especificada
		existingImageFile, err := os.Open(filename)
		handleError(err, "No se ha podido abrir el archivo: "+filename, 4)
		defer existingImageFile.Close()

		sendImage(conn, existingImageFile)

		newFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + outExtension
		// Creamos el archivo en el que se guardará la imagen convertida
		newImageFile, err := os.Create(newFilename)
		handleError(err, "No se ha podido crear el archivo de la imagen convertida", 4)
		defer newImageFile.Close()

		// Recibimos los bytes de la imagen del servidor
		newImage := receiveImage(conn)

		// Escribimos los bytes en el archivo que creamos antes
		newImageFile.Write(newImage.Bytes())

		fmt.Println("IMG_O ¡Imagen convertida recibida!")
	}

}
