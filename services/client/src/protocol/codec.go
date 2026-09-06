package protocol

import "fmt"

/* 
 * Serializa un Header en una secuencia de bytes
 *
 * El header tiene un tamaño fijo de 5 bytes:
 *
 * El primer byte para el tipo de mensaje.
 * Los bytes 1 a 4 representan el tamaño del payload.
 */     
func EncodeHeader(header Header) ([]byte, error) {
	if err := ValidateHeader(header); err != nil {
		return nil, err
	}

	buffer := make([]byte, HeaderSize)

	// Obtengo el primer byte 
	buffer[0] = byte(header.Type)

	// Utilizo shifts para obtener cada uno de los 4 bytes
	buffer[1] = byte(header.PayloadLength >> 24)
	buffer[2] = byte(header.PayloadLength >> 16)
	buffer[3] = byte(header.PayloadLength >> 8)
	buffer[4] = byte(header.PayloadLength)

	return buffer, nil
}

/*
 * Deserializa una secuencia de bytes y reconstruye
 * el header correspondiente.
 */
func DecodeHeader(data []byte) (Header, error) {
	if len(data) != HeaderSize {
		return Header{}, fmt.Errorf(
			"invalid header size: expected %d, received %d",
			HeaderSize,
			len(data),
		)
	}

	// Obtengo el tipo de mensaje.
	messageType := MessageType(data[0])

	// Ahora obtengo los bytes del tamaño del payload
	//
	// Cada byte se convierte primero a uint32 y luego se desplaza
	// hasta la posición original
	payloadLength :=
		uint32(data[1])<<24 |
			uint32(data[2])<<16 |
			uint32(data[3])<<8 |
			uint32(data[4])

	header := Header{
		Type: messageType,
		PayloadLength: payloadLength,
	}

	if err := ValidateHeader(header); err != nil {
		return Header{}, err
	}

	return header, nil
}