import os
import socket
import logger

from lottery import Lottery
from protocol.bet_codec import decode_bet, encode_bets
from protocol.connection import receive_message, send_message, send_error
from protocol.message import MessageType

BETS_FILE_NAME = "bets.csv"


class Server:
    def __init__(self, server_host: str, server_port: int, storage_dir: str) -> None:
        self.server_host = server_host
        self.server_port = server_port

        # Creo el directorio para las bets de lottery
        os.makedirs(storage_dir, exist_ok=True)

        storage_path = os.path.join(storage_dir,BETS_FILE_NAME)

        self.lottery = Lottery(storage_path)

    """
    Atiende los mensajes recibidos de un cliente
    """
    def _handle_client(self, client_socket):
        action = "handle-client"
        message_amount = 0
        agency_id = None

        try:
            logger.info(action, logger.LogResult.in_progress)

            while True:
                client_message = receive_message(client_socket)

                if client_message.header.message_type == MessageType.BET:
                    bet = decode_bet(client_message.payload)

                    # La primera bet recibida identifica el id de la agencia
                    if agency_id is None:
                        agency_id = bet.agency_id

                    lottery_bets = [bet]

                    self.lottery.store_bets(lottery_bets)

                    message_amount += 1

                    continue

                if client_message.header.message_type == MessageType.END_BETS:
                    winners = []

                    if agency_id is not None:
                        # Filtro por agency_id para devolver solamente resultados pertenecientes a esa agencia
                        winners = [
                            bet
                            for bet in self.lottery.load_bets()
                            if bet.agency_id == agency_id
                            and self.lottery.has_won(bet)
                        ]

                    payload = encode_bets(winners)

                    send_message(client_socket, MessageType.RESULTS, payload)

                    logger.info(
                        action,
                        logger.LogResult.success,
                        "messages-amount",
                        message_amount,
                    )

                    return

                if client_message.header.message_type == MessageType.ERROR:
                    client_error = client_message.payload.decode(
                        "utf-8",
                        errors="replace",
                    )

                    logger.error(
                        action,
                        logger.LogResult.fail,
                        "client-error",
                        client_error,
                    )

                    return

                raise ValueError(
                    f"unexpected message type: "
                    f"{client_message.header.message_type}"
                )

        except Exception as e:
            logger.error(
                action,
                logger.LogResult.fail,
                "messages-amount",
                message_amount,
            )

            error_message = str(e) or "server error"

            try:
                send_error(client_socket,error_message)
            except Exception:
                logger.error(
                    "send-error",
                    logger.LogResult.fail,
                )

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                with client_socket:
                    self._handle_client(client_socket)
