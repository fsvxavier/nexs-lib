package strutl

import (
	"crypto/rand"
	"math/big"
)

var randReader = rand.Reader

// Random cria uma nova string baseada em strSet. Utiliza crypto/rand como
// gerador de números aleatórios. O erro retornado é o mesmo que rand.Int retorna.
//
// Exemplos:
//
//	Random("abcdef", 5) -> "fbcae"
//	Random("123456", 3) -> "425"
func Random(strSet string, length int) (string, error) {
	if length == 0 || strSet == "" {
		return "", nil
	}
	set := []rune(strSet)
	bigLen := big.NewInt(int64(len(set)))

	res := make([]rune, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(randReader, bigLen)
		if err != nil {
			return "", err
		}
		res[i] = set[n.Int64()]
	}

	return string(res), nil
}
