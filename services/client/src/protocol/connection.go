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

/*
 * Se encargar de recibir un mensaje completo utilizando el protocolo definido
 *
 * Primero recibe los HeaderSize bytes del header 
 * Deserializa, y a partir de eso sabe cuantos bytes debe recibir del payload
 *
 * SendAll garantiza el envio de todos los bytes
 */
func ReceiveMessage(reader io.Reader) (Message, error) {

	headerBytes, err := safe_socket.RecvAll(reader, HeaderSize)
	if err != nil {
		return Message{}, err
	}

	// Deserializo el header para conocer el tipo y tamaño del payload.
	header, err := DecodeHeader(headerBytes)
	if err != nil {
		return Message{}, err
	}

	payload := []byte{}

	// Recibo exactamente la cantidad de bytes indicada en el header
	if header.PayloadLength > 0 {
		payload, err = safe_socket.RecvAll(
			reader,
			int(header.PayloadLength),
		)

		if err != nil {
			return Message{}, err
		}
	}

	message := Message{
		Header:  header,
		Payload: payload,
	}

	if err := ValidateMessage(message); err != nil {
		return Message{}, err
	}

	return message, nil
}