package protocol

import (
	"fmt"
	"strconv" // https://pkg.go.dev/strconv
	"strings" // https://pkg.go.dev/strings

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
)

// Utilizo este separador para separar los campos de una bet dentro del payload del mensaje
const (
	fieldSeparator = "|" 
	betSeparator   = "\n"
)

/*
 * Serializa una bet utilizando el formato esperado
 */
func EncodeBet(bet model.Bet) ([]byte, error) {
	if hasReservedSeparator(bet.AgencyId) || hasReservedSeparator(bet.FirstName) || hasReservedSeparator(bet.LastName) || hasReservedSeparator(bet.Birthdate) {
		return nil, fmt.Errorf("bet contains a reserved separator")
	}

	// Convierto todos los campos a string
	fields := []string{
		bet.AgencyId,
		bet.FirstName,
		bet.LastName,
		strconv.FormatUint(bet.Id, 10),
		bet.Birthdate,
		strconv.FormatUint(uint64(bet.Number), 10),
	}

	// Uno todos los campos en un solo string, separados por el separador 
	payload := []byte(strings.Join(fields, fieldSeparator))

	if len(payload) > MaxPayloadSize {
		return nil, fmt.Errorf("bet exceeds maximum payload size")
	}

	return payload, nil
}

/*
 * Deserializa las bets recibidas dentro de un mensaje de resultados
 *
 * Un payload vacio representa que no hubo ganadores
 */
func DecodeBets(payload []byte) ([]model.Bet, error) {
	// No hubo ganadores, el payload es vacio
	if len(payload) == 0 {
		return nil, nil
	}

	// Pueden haber varias bets ganadoras
	encodedBets := strings.Split(string(payload), betSeparator)
	// Creo un slice para almacenar las bets (bets, vacio, largo de la cantidad de apuestas)
	bets := make([]model.Bet, 0, len(encodedBets))

	// Deserializo cada apuesta individual
	for _, encodedBet := range encodedBets {
		// Decodifico cada bet
		bet, err := decodeBet(encodedBet)
		if err != nil {
			return nil, err
		}

		bets = append(bets, bet)
	}

	return bets, nil
}

/*
 * Deserializa una bet individual.
 */
func decodeBet(data string) (model.Bet, error) {
	// Separo los campos usando el separador
	fields := strings.Split(data, fieldSeparator)

	if len(fields) != 6 {
		return model.Bet{}, fmt.Errorf(
			"invalid bet: expected 6 fields, received %d",
			len(fields),
		)
	}

	// Campo, base 10, y 64 bits
	id, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return model.Bet{}, fmt.Errorf("invalid Id: %w", err)
	}

	number, err := strconv.ParseUint(fields[5], 10, 32)
	if err != nil {
		return model.Bet{}, fmt.Errorf("invalid bet number: %w", err)
	}

	return model.Bet{
		AgencyId:  fields[0],
		FirstName: fields[1],
		LastName:  fields[2],
		Id:  id,
		Birthdate: fields[4],
		Number:    uint32(number),
	}, nil
}

/*
 * Verifica que un campo no utilice los separadores
 */
func hasReservedSeparator(value string) bool {
	return strings.Contains(value, fieldSeparator) ||
		strings.Contains(value, betSeparator)
}