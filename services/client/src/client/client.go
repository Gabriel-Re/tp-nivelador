package client

import (
	"net"
	"time"
	"os"
	"bufio"
	"errors"
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200
const ECHO_MESSAGE_SIZE = 1024

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string

	InputFile string
	OutputFile string
}

type Client struct {
	conn   net.Conn
	config ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config}
	return client, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}

		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

/*
 * La funcion se encarga del flujo principal del cliente
 *
 * Se encarga de:
 * 1. Abrir el archivo de input.
 * 2. Crear el archivo donde se guardan las respuestas.
 * 3. Delega el procesamiento de las apuestas.
 * 4. Cierra los archivos y la conexion cuando termina.
 *
 */
func (client *Client) Run() error {
	//Me aseguro de cerrar la conexion
    defer client.conn.Close()

	//Abro el csv input
    inputFile, err := os.Open(client.config.InputFile)
    if err != nil {
        return err
    }
	//Me aseguro de cerrar el archivo input
    defer inputFile.Close()

	//Creo el output
    outputFile, err := os.Create(client.config.OutputFile)
    if err != nil {
        return err
    }
	//Me aseguro de cerrar el archivo output
    defer outputFile.Close()

    return client.processInputFile(inputFile, outputFile)
}


/*
 * La funcion se encarga de procesar el archivo linea por linea
 *
 * Cada linea representa una apuesta, para cada una:
 * 1. Lee la linea y la envia al servidor.
 * 2. Recibe la respuesta del servidor.
 * 3. Escribe esa respuesta en el archivo de salida
 *
 */
func (client *Client) processInputFile(
    inputFile *os.File,
    outputFile *os.File,
) error {

	//Uso scanner para recorrer linea por linea, por defecto usa ScanLines (https://pkg.go.dev/bufio#NewScanner)
    scanner := bufio.NewScanner(inputFile)

    for scanner.Scan() {
		//ScanLines no devuelve el fin de linea, entonces lo agrego
        message := scanner.Text() + "\n"

		//Envio al server y espero respuesta
        response, err := client.sendMessage(message)
        if err != nil {
            return err
        }

		//Escribo la respuesta en el archivo de salida verificando que hayan llegado todos los bytes
        if err := writeAllBytes(outputFile, response); err != nil {
            return err
        }
    }

	//Si scanner termino por un error se devuelve, caso de que haya leido todo el archivo devuelve nil.
    return scanner.Err()
}


/*
 * La funcion se encarga de la comunicación con el servidor.
 *
 * Recibe un mensaje, lo envia a traves del socket y esperar
 * recibir una respuesta del servidor del mismo tamaño. 
 *
 */
func (client *Client) sendMessage(message string) ([]byte, error) {

	messageBytes := []byte(message)

	// Verifico que el mensaje entre en el tamaño fijo del echo server (1024 bytes)
	if len(messageBytes) > ECHO_MESSAGE_SIZE {
		return nil, errors.New("message exceeds maximum echo message size")
	}

	// Hago un buffer fijo de 1024 bytes, los bytes sin datos los dejo en 0
	buffer := make([]byte, ECHO_MESSAGE_SIZE)

	// Copio el mensaje al comienzo del buffer
	copy(buffer, messageBytes)

	// Envio exactamente 1024 bytes
    if err := safe_socket.SendAll(
        client.conn,
        buffer,
    ); err != nil {
        return nil, err
    }

	// Como el servidor devuelve los 1024 bytes, espero exactamente esa cantidad
    response, err := safe_socket.RecvAll(
        client.conn,
        ECHO_MESSAGE_SIZE,
    )
    if err != nil {
        return nil, err
    }

	// EL servidor devolvio el pending, asi que retorno solamente los bytes correspondientes al mensaje original
    return response[:len(messageBytes)], nil
}


/*
 * Escribe todos los bytes de data usando el writer
 *
 * Como Write() puede escribir menos bytes que los solicitados,
 * se realizan escrituras sucesivas hasta completar todos los datos.
 *
 */
func writeAllBytes(writer io.Writer, data []byte) error {
	// Cantidad total de bytes escritos
	totalWritten := 0

	// Mientras queden bytes por escribir, sigo intentando
	for totalWritten < len(data) {
		
		// Escribo la parte que todavia no fue escrita
		n, err := writer.Write(data[totalWritten:])

		if err != nil {
			return err
		}

		// Acumulo los bytes escritos
		if n > 0 {
			totalWritten += n
		}

		// No hubo error pero tampoco se escribieron bytes,
		// corto para evitar loop
		if n == 0 {
			return io.ErrNoProgress
		}
	}

	return nil
}