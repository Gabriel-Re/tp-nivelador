package client

import (
	"net"
	"time"
	"os"
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200
const AMOUNT_OF_FIELDS_IN_BET = 5

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
        return client.reportError(err)
    }
	//Me aseguro de cerrar el archivo input
    defer inputFile.Close()

	//Creo el output
    outputFile, err := os.Create(client.config.OutputFile)
    if err != nil {
        return client.reportError(err)
    }
	//Me aseguro de cerrar el archivo output
    defer outputFile.Close()

    return client.processInputFile(inputFile, outputFile)
}

/*
 * Construye una Bet a partir de una linea del archivo de entrada
 * El AgencyId se recibe desde la configuracion del cliente
 */
func parseBet(line string, agencyId string) (model.Bet, error) {
	// Separo los campos de la apuesta. Esto los deja como strings asi que tengo que convertir los campos uint
	fields := strings.Split(line, ",")

	if len(fields) != AMOUNT_OF_FIELDS_IN_BET {
		return model.Bet{}, fmt.Errorf("invalid bet: expected %d fields, received %d", AMOUNT_OF_FIELDS_IN_BET, len(fields))
	}

	// Convierto el id a uint64
	id, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return model.Bet{}, fmt.Errorf(
			"invalid Id: %w",
			err,
		)
	}

	// Convierto el numero apostado a uint32
	number, err := strconv.ParseUint(fields[4], 10, 32)
	if err != nil {
		return model.Bet{}, fmt.Errorf("invalid bet number: %w", err)
	}

	return model.Bet{
		AgencyId:  agencyId,
		FirstName: fields[0],
		LastName:  fields[1],
		Id:        id,
		Birthdate: fields[3],
		Number:    uint32(number),
	}, nil
}

/*
 * La funcion se encarga de procesar el archivo linea por linea
 *
 * Cada linea representa una apuesta:
 * 1. Construye una Bet a partir de la linea leida.
 * 2. Serializa la apuesta y la envia al servidor.
 * 3. Cuando termina el archivo envia END_BETS.
 * 4. Espera el mensaje RESULTS con las apuestas ganadoras.
 * 5. Escribe los ganadores en el archivo de salida.
 */
func (client *Client) processInputFile(
	inputFile *os.File,
	outputFile *os.File,
) error {

	// Uso scanner para recorrer linea por linea, por defecto usa ScanLines (https://pkg.go.dev/bufio#NewScanner)
    scanner := bufio.NewScanner(inputFile)

	for scanner.Scan() {
		// Ignoro en caso de que haya lineas vacias
		if scanner.Text() == "" {
			continue
		}

		// Convierto la linea leida en una Bet (pkg.go.dev/bufio#Scanner.Text)
		bet, err := parseBet(
			scanner.Text(),
			client.config.AgencyId,
		)
		if err != nil {
			return client.reportError(err)
		}

		// Serializo la bet al formato definido
		payload, err := protocol.EncodeBet(bet)
		if err != nil {
			return client.reportError(err)
		}

		// Envio la bet al servidor
		if err := protocol.SendMessage(
			client.conn,
			protocol.MessageBet,
			payload,
		); err != nil {
			return err
		}
	}

	// Si scanner termino por un error lo devuelvo
	if err := scanner.Err(); err != nil {
		return client.reportError(err)
	}

	// Aviso al servidor que termine de enviar todas las bets
	if err := protocol.SendMessage(
		client.conn,
		protocol.MessageEndBets,
		nil,
	); err != nil {
		return err
	}

	// Espero el mensaje con el resultado del sorteo
	response, err := protocol.ReceiveMessage(client.conn)
	if err != nil {
		return client.reportError(err)
	}

	if response.Header.Type == protocol.MessageError {
		return fmt.Errorf("server error: %s", string(response.Payload))
	}

	if response.Header.Type != protocol.MessageResults {
		return fmt.Errorf("unexpected message type: %d", response.Header.Type)
	}

	// Deserializo las bet ganadoras
	winners, err := protocol.DecodeBets(response.Payload)
	if err != nil {
		return err
	}

	// Escribo cada ganador respetando el formato
	for _, winner := range winners {
		line := fmt.Sprintf(
			"%s,%s,%d,%s,%d\n",
			winner.FirstName,
			winner.LastName,
			winner.Id,
			winner.Birthdate,
			winner.Number,
		)

		if err := writeAllBytes(
			outputFile,
			[]byte(line),
		); err != nil {
			return err
		}
	}
	return nil
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

/*
 * Informa al servidor un error producido en el cliente
 */
func (client *Client) reportError(err error) error {
	if sendErr := protocol.SendError(client.conn,err.Error()); sendErr != nil {
		logger.Warn("send-error",logger.Fail)
	}

	return err
}