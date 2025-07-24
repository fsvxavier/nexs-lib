package ulid

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

const (
	LEN26             = 26
	LEN16             = 16
	INVALID_UNIX_YEAR = 1969
)

type UlidData struct {
	Timestamp  time.Time
	Value      string
	HexValue   string
	UUIDString string
	HexBytes   []byte
}

var (
	uld       UlidData
	errParse  error
	errDecode error
	errBytes  error
	hexValue  []byte
	uldd      ulid.ULID
	uid       uuid.UUID
	mtx       sync.RWMutex
)

func dataFromUlid(ulidID ulid.ULID) *UlidData {
	mtx.Lock()
	defer mtx.Unlock()
	uld.Timestamp = time.UnixMilli(int64(ulidID.Time()))
	uld.Value = ulidID.String()

	uld.HexValue = hex.EncodeToString(ulidID.Bytes())
	uld.HexBytes, errDecode = hex.DecodeString(uld.HexValue)
	if errDecode != nil {
		fmt.Println(errDecode.Error())
	}
	uid, errBytes = uuid.FromBytes(uld.HexBytes)
	if errBytes != nil {
		fmt.Println(errBytes.Error())
	}

	uld.UUIDString = uid.String()

	return &uld
}

// New generates a new ulid.
func NewUlid() *UlidData {
	id := ulid.Make()
	d := dataFromUlid(id)

	return d
}

// Parse tries to parses a base32 or hex uuid into ulid data.
func Parse(str string) (parsed *UlidData, err error) {
	if len(str) != LEN26 {
		str = strings.ReplaceAll(str, "-", "")
		hexValue, errDecode = hex.DecodeString(str)
		if errDecode != nil {
			return nil, errDecode
		}

		if len(hexValue) != LEN16 {
			return nil, errors.New("invalid uuid")
		}
		copy(uldd[:], hexValue)
	} else {
		uldd, errParse = ulid.Parse(str)
		if errParse != nil {
			return nil, errParse
		}
	}

	d := dataFromUlid(uldd)

	return d, nil
}

func ExtractTimestampFromUlid(uid string) (timesUlid time.Time, err error) {
	uuidId, err := uuid.Parse(uid)
	if err != nil {
		return time.Time{}, err
	}
	timestamp := uint64(uuidId[0])<<40 |
		uint64(uuidId[1])<<32 |
		uint64(uuidId[2])<<24 |
		uint64(uuidId[3])<<16 |
		uint64(uuidId[4])<<8 |
		uint64(uuidId[5])
	timesUlid = time.Unix(0, int64(timestamp)*int64(time.Millisecond))
	return timesUlid, nil
}

func IsValidUlid(value string) bool {
	errValid := uuid.Validate(value)
	if errValid != nil {
		return false
	}

	valueArray := strings.Split(value, "-")
	timestampInt, _ := strconv.ParseInt(valueArray[0]+valueArray[1], 16, 64)
	tm := time.UnixMilli(timestampInt)

	return tm.Year() != INVALID_UNIX_YEAR
}
