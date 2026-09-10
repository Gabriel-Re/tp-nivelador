package safe_socket

import (
	"io"
	"time"
)

const MAX_NO_PROGRESS_ATTEMPTS = 3
const WRITE_NO_PROGRESS_TIMEOUT = time.Second
const WRITE_RETRY_DELAY = time.Millisecond

/*
 * Envía todos los bytes recibidos a través del socket.
 * Como socket.Write() puede devolver menos bytes de los solicitados,
 * incluso cuando quedan datos por enviar, para esto realizo escrituras sucesivas
 * hasta completar el buffer.
 * Si Write() no avanza, reintento hasta WRITE_NO_PROGRESS_TIMEOUT sin progreso.
 */
func SendAll(socket io.Writer, bytes []byte) error {
	totalSent := 0
	var noProgressSince time.Time

	for totalSent < len(bytes) {
		// Envio solamente la parte del mensaje que no fue enviada
		n, err := socket.Write(bytes[totalSent:])
		
		// Los bytes hasta n fueron enviados, los acumulo
		if n > 0 {
			totalSent += n

			// Reinicio el tiempo sin avance
			noProgressSince = time.Time{}
		}
		
		if err != nil {
			return err
		}
		
		// Un Write sin avance puede ser transitorio. Achico la espera
		if n == 0 {
			if noProgressSince.IsZero() {
				noProgressSince = time.Now()
			}
			if time.Since(noProgressSince) >= WRITE_NO_PROGRESS_TIMEOUT {
				return io.ErrNoProgress
			}
			time.Sleep(WRITE_RETRY_DELAY)
		}
	}
	return nil
}

/*
 * Recibe exactamente el 'size' bytes desde el socket
 * Como socket.Read() puede devolver menos bytes de los solicitados,
 * incluso cuando quedan datos por recibir, para esto realizo lecturas sucesivas
 * hasta completar el buffer.
 * En caso de que Read() no avance, se realizan un máximo de MAX_NO_PROGRESS_ATTEMPTS intentos antes de retornar error.
 */
func RecvAll(socket io.Reader, size int) ([]byte, error) {

	buff := make([]byte, size)

	totalReceived := 0
	noProgressAttempts := 0

	for totalReceived < size {
		// Leo solamente sobre la parte del buffer que falta completar.
		n, err := socket.Read(buff[totalReceived:])

		// Acumulo los bytes recibidos
		if n > 0 {
			totalReceived += n

			// Reinicio el contador
			noProgressAttempts = 0
		}

		// Si ya recibi todos los bytes, retorno
		if totalReceived == size {
			return buff, nil
		}

		if err != nil {
			return nil, err
		}

		// Si Read no devuelve bytes ni error, significa que no avanzó.
		// Corto para evitar loop infinito.
		if n == 0 {
			noProgressAttempts++

			if noProgressAttempts >= MAX_NO_PROGRESS_ATTEMPTS {
				return nil, io.ErrNoProgress
			}
		}
	}
	return buff, nil
}
