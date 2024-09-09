package kennitala

import (
	"fmt"
	"time"

	kennitalaerrors "github.com/noona-hq/kennitala/kennitalaerror"
	utils "github.com/noona-hq/kennitala/utils"
)

var (
	ErrInvalidKennitalaType        = errInvalidKennitalaType()
	ErrInvalidKennitalaLength      = errInvalidKennitalaLength()
	ErrInvalidKennitalaCentury     = errInvalidKennitalaCentury()
	ErrInvalidKennitalaFirstLetter = errInvalidKennitalaFirstLetter()
	ErrInvalidKennitalaCheckDigit  = errInvalidKennitalaCheckDigit()
	ErrInvalidKennitalaDate        = errInvalidKennitalaDate()
)

func errInvalidKennitalaType() error        { return kennitalaerrors.ErrInvalidKennitalaType }
func errInvalidKennitalaLength() error      { return kennitalaerrors.ErrInvalidKennitalaLength }
func errInvalidKennitalaCentury() error     { return kennitalaerrors.ErrInvalidKennitalaCentury }
func errInvalidKennitalaFirstLetter() error { return kennitalaerrors.ErrInvalidKennitalaFirstLetter }
func errInvalidKennitalaCheckDigit() error  { return kennitalaerrors.ErrInvalidKennitalaCheckDigit }
func errInvalidKennitalaDate() error        { return fmt.Errorf("invalid birthdate in kennitala") }

type Kennitala string

type KennitalaType int8

const (
	KennitalaIndividual KennitalaType = 1 << iota
	KennitalaCompany
	KennitalaSystem
	KennitalaAllTypes KennitalaType = KennitalaIndividual | KennitalaCompany | KennitalaSystem
)

func (kennitalaType KennitalaType) isValidKennitalaType() error {
	switch kennitalaType {
	case KennitalaIndividual, KennitalaCompany, KennitalaSystem, KennitalaAllTypes:
		return nil
	}
	return errInvalidKennitalaType()
}

func (kennitalaType KennitalaType) hasFlag(flag KennitalaType) bool { return kennitalaType&flag != 0 }

func (kennitala Kennitala) IsValidKennitala(kennitalaType KennitalaType) error {
	if err := kennitalaType.isValidKennitalaType(); err != nil {
		return err
	}

	if len(kennitala) != 10 {
		return errInvalidKennitalaLength()
	}

	// Validate century and date
	if err := kennitala.validateBirthdateAndCentury(); err != nil {
		return err
	}

	allowFirstLetters := map[string]string{}
	if kennitalaType.hasFlag(KennitalaIndividual) {
		// Kennitala for individuals starts with 0, 1, 2 and 3
		allowFirstLetters["0"] = "0"
		allowFirstLetters["1"] = "1"
		allowFirstLetters["2"] = "2"
		allowFirstLetters["3"] = "3"
	}
	if kennitalaType.hasFlag(KennitalaCompany) {
		// Kennitala for companies starts with 4, 5, 6 and 7
		allowFirstLetters["4"] = "4"
		allowFirstLetters["5"] = "5"
		allowFirstLetters["6"] = "6"
		allowFirstLetters["7"] = "7"
	}
	if kennitalaType.hasFlag(KennitalaSystem) {
		// Kerfiskennitala start with 8 and 9
		allowFirstLetters["8"] = "8"
		allowFirstLetters["9"] = "8"
	}

	first := string(kennitala[0])
	_, exists := allowFirstLetters[first]

	if !exists {
		return errInvalidKennitalaFirstLetter()
	}

	// Validate check digit
	checkDigit, _ := utils.StringToInt(string(kennitala[8]))
	calculatedCheckDigit, _ := calculateCheckDigit(kennitala)

	if checkDigit != calculatedCheckDigit {
		return errInvalidKennitalaCheckDigit()
	}

	return nil
}

// validateBirthdateAndCentury validates that the birthdate corresponds to the century
func (kennitala Kennitala) validateBirthdateAndCentury() error {
	// Extract the birth date
	day := kennitala[:2]
	month := kennitala[2:4]
	year := kennitala[4:6]
	centuryDigit := kennitala[9]

	// Parse the year based on the century indicated by the ninth digit
	var fullYear string
	switch centuryDigit {
	case '8':
		fullYear = "18" + string(year)
	case '9':
		fullYear = "19" + string(year)
	case '0':
		fullYear = "20" + string(year)
	default:
		return errInvalidKennitalaCentury()
	}

	// Try to parse the birth date as a valid date
	birthDate := fmt.Sprintf("%s-%s-%s", fullYear, month, day)
	_, err := time.Parse("2006-01-02", birthDate)
	if err != nil {
		return errInvalidKennitalaDate()
	}

	return nil
}

func (kennitala Kennitala) IsPerson() error {
	return kennitala.IsValidKennitala(KennitalaIndividual)
}

func calculateCheckDigit(kennitala Kennitala) (int8, error) {
	if len(kennitala) != 10 {
		return -1, errInvalidKennitalaLength()
	}

	multiples := [8]int8{3, 2, 7, 6, 5, 4, 3, 2}

	sum := uint16(0)
	for i := uint8(0); i < 8; i++ {
		num, _ := utils.StringToInt(string(kennitala[i]))
		sum += uint16(num * multiples[i])
	}

	parity := (sum % 11)
	if parity == 0 {
		return 0, nil
	}
	parity = 11 - parity
	if parity == 10 {
		return 0, errInvalidKennitalaCheckDigit()
	}

	return int8(parity), nil
}
