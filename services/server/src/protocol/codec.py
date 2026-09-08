from .message import (
    HEADER_SIZE,
    MAX_PAYLOAD_SIZE,
    Header,
    MessageType,
)

"""Valida que el header respete las restricciones del protocolo"""
def validate_header(header: Header) -> None:
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