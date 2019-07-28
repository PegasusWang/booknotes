// base62 algorithm

package base62

import (
	"log"
	"math"
	"strings"
)

/*

def get_table():
	table1 = [chr(ord('a') + i) for i in range(26)]
	table2 = [chr(ord('A') + i) for i in range(26)]
	table3 = [str(i) for i in range(10)]
	table = table1 + table2 + table3  # 62 进制表
	return table

*/

const base = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const b = 62

// ToBase62 encodes the given ID to a base62 string
func ToBase62(num int) string {
	r := num % b
	res := string(base[r])
	div := num / b
	q := int(math.Floor(float64(div)))
	for q != 0 {
		r = q % b
		temp := q / b
		q = int(math.Floor(float64(temp)))
		res = string(base[int(r)]) + res
	}
	return string(res)
}

// ToBase10 decodes a given base62 string to datbase ID
func ToBase10(str string) int {
	res := 0
	for _, r := range str {
		res = (b * res) + strings.Index(base, string(r))
	}
	return res
}

func test() {
	x := 100
	base62String := base62.ToBase62(x)
	log.Println(base62String)
	normalNumber := base62.ToBase10(base62String)
	log.Println(normalNumber)
}
