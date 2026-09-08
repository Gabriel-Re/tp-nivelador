import socket

import safe_socket

from .codec import decode_header, encode_header, validate_message
from .message import HEADER_SIZE, Header, Message, MessageType


"""
Primero recibe el header de tamaño fijo y luego utiliza el tamaño
indicado en el header para recibir exactamente los bytes del payload
"""
def receive_message(socket: socket.socket) -> Message:
    header_bytes = safe_socket.recv_all(
        socket,
        HEADER_SIZE,
    )

    if not header_bytes:
        raise ConnectionError("socket closed before receiving message header")

    # Deserializo el header para conocer el tipo y tamaño del payload
    header = decode_header(header_bytes)

    payload = b""

    if header.payload_length > 0:
        payload = safe_socket.recv_all(
            socket,
            header.payload_length,
        )

        if len(payload) != header.payload_length:
            raise ConnectionError("socket closed before receiving complete payload")

    message = Message(header=header,payload=payload)

    validate_message(message)

    return message


"""
Envia un mensaje completo utilizando el protocolo definido.

Primero serializa y envia el header y luego envia el payload
solamente cuando el mensaje contiene datos.
"""
def send_message(socket: socket.socket,message_type: MessageType,payload: bytes = b"") -> None:
    header = Header(
        message_type=message_type,
        payload_length=len(payload),
    )

    message = Message(
        header=header,
        payload=payload,
    )

    validate_message(message)

    # Serializo y envio el header
    header_bytes = encode_header(header)

    safe_socket.send_all(
        socket,
        header_bytes,
    )

    # Algunos tipos de mensajes pueden no tener payload
    if payload:
        safe_socket.send_all(
            socket,
            payload,
        )