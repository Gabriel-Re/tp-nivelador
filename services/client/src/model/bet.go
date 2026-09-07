package protocol

/*
 * Representa los datos transportados dentro
 * del payload de un mensaje BET o RESULTS.
 *
 * Ej: {8,Santiago Lionel,Lorca,30904465,1999-03-17,7574}
 */
type BetPayload struct {
	AgencyId  uint32
	FirstName string
	LastName  string
	Id  uint64
	Birthdate string
	Number    uint32
}