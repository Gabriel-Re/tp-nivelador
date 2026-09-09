from lottery import Bet


FIELD_SEPARATOR = "|"
BET_SEPARATOR = b"\n"
AMOUNT_OF_FIELDS_IN_BET = 6


"""
Deserializa una bet dando el formato esperado por Lottery

Ej: Recibo "0|Santiago Lionel|Lorca|30904465|1999-03-17|7574"
"""
def decode_bet(payload: bytes) -> Bet:
    try:
        fields = payload.decode("utf-8").split(FIELD_SEPARATOR)
    except UnicodeDecodeError as error:
        raise ValueError("invalid bet encoding")

    # Si no cumpe con el largo de campos esperados retorno error
    if len(fields) != AMOUNT_OF_FIELDS_IN_BET:
        raise ValueError(f"invalid bet: expected {AMOUNT_OF_FIELDS_IN_BET} fields, " f"received {len(fields)}")

    try:
        agency_id = int(fields[0])
        document = int(fields[3])
        number = int(fields[5])
    except ValueError as error:
        raise ValueError("invalid numeric bet field")

    return Bet(
        agency_id = agency_id,
        first_name = fields[1],
        last_name = fields[2],
        document = document,
        birthdate = fields[4],
        number = number,
    )


"""
Deserializa un batch de bets

Cada bet se encuentra separada mediante un salto de linea.
"""
def decode_bets(payload: bytes) -> list[Bet]:
    if not payload:
        raise ValueError("empty bet batch")

    # Separo las distintas bets
    encoded_bets = payload.split(BET_SEPARATOR)

    bets = []

    for encoded_bet in encoded_bets:
        bet = decode_bet(encoded_bet)
        bets.append(bet)

    return bets


"""
Serializa una bet individual
"""
def encode_bet(bet: Bet) -> bytes:
    fields = [
        str(bet.agency_id),
        bet.first_name,
        bet.last_name,
        str(bet.document),
        bet.birthdate,
        str(bet.number),
    ]

    return FIELD_SEPARATOR.join(fields).encode("utf-8")


"""
Serializa las bets ganadoras dentro del payload de RESULTS.

Cada bet se separa mediante un salto de linea
"""
def encode_bets(bets: list[Bet]) -> bytes:
    return BET_SEPARATOR.join(
        encode_bet(bet) for bet in bets
    )