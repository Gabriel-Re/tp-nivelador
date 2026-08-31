package safe_socket

import "io"


/*
 * Envía todos los bytes recibidos a través del socket.
 * Como socket.Write() puede devolver menos bytes de los solicitados,
 * incluso cuando quedan datos por enviar, para esto realizo escrituras sucesivas
 * hasta completar el buffer.
 */
func SendAll(socket io.Writer, bytes []byte) error {
	totalSent := 0

	for totalSent < len(bytes) {
		// Envio solamente la parte del mensaje que no fue enviada
		n, err := socket.Write(bytes[totalSent:])
		if err != nil {
			return err
		}
		
		// Evito quedar en loop infinito si Write no avanza
		if n == 0 {
			return io.ErrShortWrite
		}
		
		totalSent += n
	}
	return nil
}

/*
 * Recibe exactamente el 'size' bytes desde el socket
 * Como socket.Read() puede devolver menos bytes de los solicitados,
 * incluso cuando quedan datos por recibir, para esto realizo lecturas sucesivas
 * hasta completar el buffer.
 */
func RecvAll(socket io.Reader, size int) ([]byte, error) {

	buff := make([]byte, size)

	totalReceived := 0
	for totalReceived < size {
		// Leo solamente sobre la parte del buffer que falta completar.
		n, err := socket.Read(buff[totalReceived:])

		if err != nil {
			return nil, err
		}

		// Si Read no devuelve bytes ni error, significa que no avanzó.
		// Corto para evitar loop infinito.
		if n == 0 {
			return nil, io.ErrNoProgress
		}

		totalReceived += n
	}
	return buff, nil
}
