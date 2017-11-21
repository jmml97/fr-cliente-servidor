package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"net"
	"os"
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

func convertJPEGToPNG(w io.Writer, r io.Reader) error {
	img, err := jpeg.Decode(r)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = png.Encode(w, img)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func convertPNGToJPEG(w io.Writer, r io.Reader) error {
	img, err := png.Decode(r)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = jpeg.Encode(w, img, nil)
	if err != nil {
		fmt.Println(err)
	}
	return err

}

func receiveImage(conn net.Conn) bytes.Buffer {

	var receivedImage bytes.Buffer

	// Recibimos primero el tamaño de la imagen
	var receivedImageSize int64
	err := binary.Read(conn, binary.LittleEndian, &receivedImageSize)
	handleError(err, "Error durante la recepción del tamaño de la imagen", 1)

	fmt.Println("Tamaño de la imagen a recibir:", receivedImageSize)

	// Ahora recibimos la imagen en sí
	n, err := io.CopyN(&receivedImage, conn, receivedImageSize)
	handleError(err, "Error durante la recepción de la imagen", 1)

	fmt.Println(n, "bytes recibidos")
	return receivedImage

}

func sendImage(conn net.Conn, imageData []byte) {

	imageSize := int64(len(imageData))

	fmt.Println("Tamaño de la imagen a enviar:", imageSize)

	// Enviamos el tamaño de la imagen
	err := binary.Write(conn, binary.LittleEndian, imageSize)
	handleError(err, "Error al enviar el tamaño de la imagen", 2)

	reader := bytes.NewReader(imageData)

	n, err := io.CopyN(conn, reader, imageSize)
	handleError(err, "Error durante el envío de la imagen", 2)

	fmt.Println(n, "bytes enviados")

}

func showMenu(conn net.Conn) {
	// \f simboliza fin del mensaje
	menuMessage := "Seleccione el tipo conversión que quiere realizar\n" +
		"1- PNG a JPEG\n" +
		"2- JPEG a PNG\n" +
		"3- Salir\n\f"
	conn.Write([]byte(menuMessage))
}

func reveiveOption(conn net.Conn) int {
	optionString, _ := bufio.NewReader(conn).ReadString('#')
	option, _ := strconv.Atoi(strings.Trim(optionString, "#"))
	fmt.Println(optionString)
	fmt.Println(option)
	return option
}

func main() {

	fmt.Println("Iniciando el servidor...")

	// listen on all interfaces
	ln, _ := net.Listen("tcp", ":8081")

	for {
		// accept connection on port
		conn, _ := ln.Accept()
		defer conn.Close()

		go func() { // goroutine concurrente

			for {

				showMenu(conn)

				option := reveiveOption(conn)

				if option == 3 {
					fmt.Println("Se termina la ejecución del cliente")
					break
				}

				fmt.Println("Comienza la recepción de la imagen")
				oldImage := receiveImage(conn)
				fmt.Println("Imagen a convertir recibida")
				reader := bytes.NewReader(oldImage.Bytes())
				var newImage bytes.Buffer
				writer := bufio.NewWriter(&newImage)

				switch option {
				case 1:
					convertPNGToJPEG(writer, reader)
				case 2:
					convertJPEGToPNG(writer, reader)
				}

				fmt.Println(len(newImage.Bytes()))

				fmt.Println("Imagen convertida")
				sendImage(conn, newImage.Bytes())
				fmt.Println("Imagen enviada al cliente")
			}

		}()

	}

}
