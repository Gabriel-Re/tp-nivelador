package client

import (
	"net"
	"time"
	"os"
	"bufio"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

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

		//Guardo el archivo de salida
        if _, err := outputFile.Write(response); err != nil {
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

	// Envía todos los bytes del mensaje al servidor
    if err := safe_socket.SendAll(
        client.conn,
        []byte(message),
    ); err != nil {
        return nil, err
    }

	// Como el servidor hace echo, esperamos recibir la misma cantidad de bytes que enviamos
    response, err := safe_socket.RecvAll(
        client.conn,
        len(message),
    )
    if err != nil {
        return nil, err
    }

    return response, nil
}