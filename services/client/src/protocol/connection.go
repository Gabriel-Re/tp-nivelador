package protocol

import (
	"io"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

/*
 * Se encarga de enviar un mensaje completo utilizando el protocolo definido
 *
 * Primero se construye y envía el header de tamaño fijo
 * Luego envía el payload solamente cuando el mensaje contiene datos
 *
 * SendAll garantiza el envio de todos los bytes
 */
func SendMessage(writer io.Writer, messageType MessageType, payload []byte,) error {

	message := Message{
		Header: Header{Type: messageType, PayloadLength: uint32(len(payload))},
		Payload: payload,
	}

	if err := ValidateMessage(message); err != nil {
		return err
	}

	// Serializo el header
	headerBytes, err := EncodeHeader(message.Header)
	if err != nil {
		return err
	}

	// Envío todos los bytes del header
	if err := safe_socket.SendAll(writer, headerBytes); err != nil {
		return err
	}

	// Algunos tipos pueden no tener payload
	if len(payload) == 0 {
		return nil
	}

	// Envío todos los bytes correspondientes al payload.
	if err := safe_socket.SendAll(writer, payload); err != nil {
		return err
	}

	return nil
}

// Receivemessage seria del mismo orden
// Recibo el header, lo decodifico y luego recibo el payload. Como recv all se encargar del short read, no hay que preocuparse por eso.
// Al final deberia valdiar si es