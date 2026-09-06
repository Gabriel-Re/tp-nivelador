package protocol

import "fmt"

type MessageType byte

/*
 * Los distintos tipos de mensajes que se pueden enviar y recibir.
 */
const (
	MessageBet MessageType = 1
	MessageEndBets MessageType = 2
	MessageResults MessageType = 3
	MessageError MessageType = 4
)

/* 
 * Tamaño del header y del payload de un mensaje.
 */
const (
	HeaderSize = 5
	MaxPayloadSize = 1024 //TODO tamaño máximo del payload (a definir) 
)

/*
 * Estructura del header del mensaje.
 */
type Header struct {
	Type MessageType
	PayloadSize uint32
}

/*
 * Estructura del mensaje que se enviará y recibirá por el socket.
 */
type Message struct {
	Header Header
	Payload []byte
}

// Funciones TODO
// Mensaje valido, Header valido, Payload valido? Serialization, Deserialization?


/*
 * Verifica si un tipo de mensaje es válido.
 */
func IsValidMessageType(messageType MessageType) bool {
	switch messageType {
	case MessageBet, MessageEndBets, MessageResults, MessageError:
		return true
	default:
		return false
	}
}

/*
 * Valida el header de un mensaje.
 */
func ValidateHeader(header Header) error {
	if !IsValidMessageType(header.Type) {
		return fmt.Errorf("unknown message type: %d", header.Type)
	}

	if header.PayloadLength > MaxPayloadSize {
		return fmt.Errorf(
			"payload too large: %d bytes",
			header.PayloadLength,
		)
	}

	return nil
}