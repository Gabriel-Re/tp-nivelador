package safe_socket

import "io"

const MAX_NO_PROGRESS_ATTEMPTS = 3

/*
 * Envía todos los bytes recibidos a través del socket.
 * Como socket.Write() puede devolver menos bytes de los solicitados,
 * incluso cuando quedan datos por enviar, para esto realizo escrituras sucesivas
 * hasta completar el buffer.
 * En caso de que Write() no avance, se realizan un máximo de MAX_NO_PROGRESS_ATTEMPTS intentos antes de retornar error.
 */
func SendAll(socket io.Writer, bytes []byte) error {
	totalSent := 0
	noProgressAttempts := 0

	for totalSent < len(bytes) {
		// Envio solamente la parte del mensaje que no fue enviada
		n, err := socket.Write(bytes[totalSent:])
		
		// Los bytes hasta n fueron enviados, los acumulo
		if n > 0 {
			totalSent += n

			// Reinicio el contador de intentos sin avance
			noProgressAttempts = 0
		}
		
		if err != nil {
			return err
		}
		
		// Si Write no envió ningun byte, vuelvo a intentar
		// Limito los intentos para evitar un posible loop infinito
		if n == 0 {
			noProgressAttempts++

			if noProgressAttempts >= MAX_NO_PROGRESS_ATTEMPTS {
				return io.ErrNoProgress
			}
		}
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
