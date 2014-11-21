// @description make uuid, support short uuid
// @reference   https://github.com/stochastic-technologies/shortuuid
// @authors     Odin

package utils

import (
    "code.google.com/p/go-uuid/uuid"
    "fmt"
    "math/big"
    "strings"
)

const (
    alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
    shortlen = 22 //这个长度是根据alphabet得来的,省略计算步骤
)

func ShortenUUID(s string) (ss string) {
    //uuInt, tr, tm, off := new(big.Int), new(big.Int), new(big.Int), new(big.Int)
    uuInt, off := new(big.Int), new(big.Int)
    //remove "-"
    s = strings.ToLower(strings.Replace(s, "-", "", -1))
    //fmt.Sscan("0x"+s, uuInt)
    uuInt.SetString(s, 16)

    alphaLen := big.NewInt(int64(len(alphabet)))
    for uuInt.Cmp(big.NewInt(0)) > 0 {
        //uuInt, off = tr.DivMod(uuInt, alphaLen, tm)
        uuInt, off = uuInt.DivMod(uuInt, alphaLen, off)
        ss += string(alphabet[off.Int64()])
    }
    //如果不足22位,用第一个字符补全
    if diff := shortlen - len(ss); diff > 0 {
        ss += strings.Repeat(string(alphabet[0]), diff)
    }

    return
}

func LengthenUUID(s string) (ls string) {
    uuInt, off := new(big.Int), new(big.Int)
    alphaLen := big.NewInt(int64(len(alphabet)))
    //需要倒序
    for i := len(s) - 1; i >= 0; i-- {
        char := s[i]
        off = big.NewInt(int64(strings.Index(alphabet, string(char))))
        uuInt = uuInt.Add(uuInt.Mul(uuInt, alphaLen), off)
    }
    //转为16进制
    if b := fmt.Sprintf("%x", uuInt); b != "" {
        //fmt.Println(b)
        //b := []byte(b)
        ll := len(b)
        //考虑到前缀为零的情况. 倒过来,先满足后面的字符位数
        ls = fmt.Sprintf("%08v-%04v-%04v-%04v-%012v",
            b[:ll-24], b[ll-24:ll-20], b[ll-20:ll-16], b[ll-16:ll-12], b[ll-12:])
    }
    return
}

//default, version 4
func NewUUID() string {
    return uuid.New()
}

func NewShortUUID() string {
    newUUID := NewUUID()
    return ShortenUUID(newUUID)
}

// generate uuid v5
// namespace直接写域名或者url
func NewUUID5(namespace, data string) string {
    //以http开头的为URL, 其余都为DNS
    if strings.HasPrefix(strings.ToLower(namespace), "http") {
        return uuid.NewSHA1(uuid.NameSpace_URL, []byte(namespace+data)).String()
    } else {
        return uuid.NewSHA1(uuid.NameSpace_DNS, []byte(namespace+data)).String()
    }
}

func NewShortUUID5(namespace, data string) string {
    newUUID := NewUUID5(namespace, data)
    return ShortenUUID(newUUID)
}
