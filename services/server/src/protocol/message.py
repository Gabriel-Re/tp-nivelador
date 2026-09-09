from dataclasses import dataclass
from enum import IntEnum

# 1 byte para el tipo de mensaje
# 4 bytes para el tamaño del payload
HEADER_SIZE = 5

MAX_PAYLOAD_SIZE = 1024 * 1024  # 1 MB

"""Tipos de mensajes definidos por el protocolo"""
class MessageType(IntEnum):
    BET = 1
    END_BETS = 2
    RESULTS = 3
    ERROR = 4
    ACK = 5


"""Información necesaria para interpretar el payload de un mensaje"""
@dataclass
class Header:
    message_type: MessageType
    payload_length: int


"""Representa un mensaje del protocolo"""
@dataclass
class Message:
    header: Header
    payload: bytes