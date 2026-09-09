import os
import sys

import logger
import server

SERVER_HOST = os.environ["SERVER_HOST"]
SERVER_PORT = int(os.environ["SERVER_PORT"])
AGENCY_QUORUM_MIN = int(os.environ["AGENCY_QUORUM_MIN"])

# En este directorio se almacenan las bets
SERVER_STORAGE_DIR = os.environ.get("SERVER_STORAGE_DIR","/data")


def main():
    logger.init()

    if AGENCY_QUORUM_MIN <= 0:
        raise ValueError("AGENCY_QUORUM_MIN must be greater than zero")

    s = server.Server(SERVER_HOST, SERVER_PORT, SERVER_STORAGE_DIR, AGENCY_QUORUM_MIN)
    try:
        s.run()
    except Exception as e:
        logger.error("server-run", logger.LogResult.fail, "err", e)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())