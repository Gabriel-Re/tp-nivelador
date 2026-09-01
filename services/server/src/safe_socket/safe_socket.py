import socket

MAX_NO_PROGRESS_ATTEMPTS = 3

"""
Recibe 'size' bytes del socket

Como recv() puede devolver menos bytes de los esperados,
se realizan lecturas sucesivas hasta completar el mensaje.
"""
def recv_all(socket: socket.socket, size):
    # Acumulo los bytes recibidos
    buffer = bytearray()

    while len(buffer) < size:
        
        remaining = size - len(buffer)

        chunk = socket.recv(remaining)

        # Si retorna vacio (b") significa que el cliente se desconecto
        if chunk == b"":
            raise ConnectionError("socket closed before receiving all expected bytes")        

        # Agrego al buffer
        buffer.extend(chunk)

    # Retorno los bytes recibidos
    return bytes(buffer)



"""
Se envia los bytes recibidos por el socket

Como send() puede enviar menos bytes de los solicitados,
se realizan envíos sucesivos hasta completar el mensaje.
"""
def send_all(sock: socket.socket, data: bytes) -> int:

    total_sent = 0
    no_progress_attempts = 0

    while total_sent < len(data):

        # Envio los bytes que faltan
        sent = sock.send(data[total_sent:])

        # Caso que no hubo envio
        if sent == 0:
            no_progress_attempts += 1

            if no_progress_attempts >= MAX_NO_PROGRESS_ATTEMPTS:
                raise ConnectionError("socket made no progress while sending data")

            continue

        # Reinicio contador en caso de prgoresar
        no_progress_attempts = 0

        # Acumulo la cantidad enviada
        total_sent += sent

    return total_sent