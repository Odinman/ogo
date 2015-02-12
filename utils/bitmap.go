//这里实现bitmap索引
package utils

import (
	"fmt"
	"math/big"
	"sort"
)

type BitmapIndex struct { // bitmap索引数据结构
	Data  []byte //数据用[]byte存放,一个元素(block)8bit
	Start int    //开始的block
	End   int    //结束的block
}

/* {{{ func ReadIntFromBytes(bs []byte) (r int)
 *
 */
func ReadIntFromBytes(bs []byte) (r int) {
	l := len(bs)
	for i, b := range bs {
		shift := uint((l - i - 1) * 8)
		r |= int(b) << shift
	}
	return
}

/* }}} */

/* {{{ func NewBitmapIndex(s []int) *BitmapIndex
 * 根据一个整数slice建立索引
 */
func NewBitmapIndex(s []int) *BitmapIndex {
	sort.Ints(s) //先排序
	bi := new(BitmapIndex)
	bi.Start = s[0] / 8
	bi.End = s[len(s)-1] / 8

	b := big.NewInt(0)
	one := big.NewInt(1)
	rcver := big.NewInt(0)

	for _, sv := range s {
		offset := sv - bi.Start*8 //差多少就有多少位
		b.Or(b, rcver.Lsh(one, uint(offset)))
	}
	bi.Data = b.Bytes()
	return bi
}

/* }}} */

/* {{{ func ReadBitmapIndex(ib []byte) (*BitmapIndex,error)
 * 从[]byte里读取BitmapIndex,应用场景为从文件或者内存中拿到[]byte,转化为索引
 */
func ReadBitmapIndex(ib []byte) (bi *BitmapIndex, err error) {
	il := len(ib)
	if il <= 8 { //至少比8大
		return nil, fmt.Errorf("can't read from %s", ib)
	}

	bi = new(BitmapIndex)
	bi.Data = ib[8:]
	start := ReadIntFromBytes(ib[:4])
	end := ReadIntFromBytes(ib[4:8])
	bi.Start = start / 8
	bi.End = end / 8

	return
}

/* }}} */

/* {{{ func (bi *BitmapIndex) Bytes() ([]byte, error)
 * 将BitmapIndex转化为[]byte,方便存放到文件或者内存中去
 */
func (bi *BitmapIndex) Bytes() (ib []byte, err error) {
	bl := len(bi.Data)
	if bl <= 0 {
		return nil, fmt.Errorf("index data empty")
	}
	start := bi.Start * 8
	end := bi.End * 8
	bs := big.NewInt(int64(start)).Bytes()
	be := big.NewInt(int64(end)).Bytes()

	ib = make([]byte, 8+bl)   //8+数据长度
	copy(ib[4-len(bs):4], bs) //1-4字节保存开始block
	copy(ib[8-len(be):8], be) // 5-8字节保存结束block
	copy(ib[8:], bi.Data)

	return
}

/* }}} */

/* {{{ func (bi *BitmapIndex) Slices() ([]int, error)
 * 将BitmapIndex转化为[]byte,方便存放到文件或者内存中去
 */
func (bi *BitmapIndex) Slices() (s []int, err error) {
	if bi == nil || len(bi.Data) <= 0 {
		return nil, fmt.Errorf("not found item")
	}
	s = make([]int, 0)
	Len := len(bi.Data)
	for i, b := range bi.Data {
		if b > 0 { //双方都大于0才有比较的意义
			for ii := 0; ii < 8; ii++ { //遍历8bit
				if b&(1<<uint(ii)) > 0 { //找到交集位置！
					shift := ii + (Len-i-1)*8 //偏移量
					s = append(s, bi.Start*8+shift)
				}
			}
		}
	}

	return
}

/* }}} */

/* {{{ func (bi *BitmapIndex) And(obi *BitmapIndex) *BitmapIndex
 * 求交集
 */
func (bi *BitmapIndex) And(obi *BitmapIndex) (nbi *BitmapIndex) {
	if bi == nil {
		return nil
	}
	if bi.End < obi.Start || obi.End < bi.Start {
		//不可能有交集
		return nil
	}
	var start, end int
	if bi.Start < obi.Start {
		//以大的start为准
		start = obi.Start
	} else {
		start = bi.Start
	}
	if bi.End < obi.End {
		//以小的end为准
		end = bi.End
	} else {
		end = obi.End
	}

	//得到两个索引的重叠部分
	data1 := bi.Data[bi.End-end : len(bi.Data)-(start-bi.Start)]
	data2 := obi.Data[obi.End-end : len(obi.Data)-(start-obi.Start)]

	nbi = new(BitmapIndex)
	nbi.Start = start
	nbi.End = end

	Len := end - start + 1
	nbi.Data = make([]byte, Len)

	for i, b1 := range data1 {
		b2 := data2[i]
		nbi.Data[i] = b1 & b2
		//if b1 > 0 && b2 > 0 { //双方都大于0才有比较的意义
		//	b3 := b1 & b2
		//	if b3 > 0 { //交集也大于0
		//		for ii := 0; ii < 8; ii++ { //遍历8bit
		//			if b3&(1<<uint(ii)) > 0 { //找到交集位置！
		//				shift := ii + (Len-i-1)*8 //偏移量
		//			}
		//		}
		//	}
		//}
	}

	return
}

/* }}} */
