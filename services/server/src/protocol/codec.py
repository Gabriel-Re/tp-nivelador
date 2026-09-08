from .message import (
    HEADER_SIZE,
    MAX_PAYLOAD_SIZE,
    Header,
    MessageType,
    Message,
)

"""Valida que el mensaje respete las restricciones del protocolo"""
def validate_message(message: Message) -> None:
    validate_header(message.header)

    if len(message.payload) != message.header.payload_length:
        raise ValueError(f"payload length mismatch: "f"header={message.header.payload_length} "f"actual={len(message.payload)}")

    if message.header.message_type == MessageType.BET:
        if not message.payload:
            raise ValueError("BET message requires a payload")

    elif message.header.message_type == MessageType.END_BETS:
        if message.payload:
            raise ValueError("END_BETS message must have empty payload")

    elif message.header.message_type == MessageType.ERROR:
        if not message.payload:
            raise ValueError("ERROR message requires a payload")

"""Valida que el header respete las restricciones del protocolo"""
def validate_header(header: Header) -> None:
    try:
        MessageType(header.message_type)
    except (ValueError, TypeError):
        raise ValueError(f"unknown message type: {header.message_type}")

    if header.payload_length < 0:
        raise ValueError("payload length cannot be negative")

    if header.payload_length > MAX_PAYLOAD_SIZE:
        raise ValueError(
            f"payload too large: {header.payload_length} bytes"
        )

"""
Serializa un Header utilizando el formato del protocolo

Primer byte = tipo de mensaje
Siguientes 4 bytes = tamaño del payload
"""
def encode_header(header: Header) -> bytes:
    validate_header(header)

    message_type = bytes([header.message_type])

    payload_length = header.payload_length.to_bytes(
        4,
        byteorder="big", # Con esto me aseguro de que sea big endian
    )

    return message_type + payload_length


"""Reconstruye un Header a partir de sus 5 bytes"""
def decode_header(data: bytes) -> Header:
    if len(data) != HEADER_SIZE:
        raise ValueError(
            f"invalid header size: expected {HEADER_SIZE}, got {len(data)}"
        )

    # Obtengo el primer byte
    try:
        message_type = MessageType(data[0])
    except ValueError as error:
        raise ValueError(f"unknown message type: {data[0]}")

    # Obtengo el largo del payload
    payload_length = int.from_bytes(
        data[1:5],
        byteorder="big", # Con esto me aseguro de que sea big endian
    )

    header = Header(
        message_type=message_type,
        payload_length=payload_length,
    )

    validate_header(header)

    return header