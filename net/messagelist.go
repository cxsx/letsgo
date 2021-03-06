/*=============================================================================
#     FileName: messagelist.go
#         Desc: Message pack/unpack
#       Author: sunminghong
#        Email: allen.fantasy@gmail.com
#     HomePage: http://weibo.com/5d13
#      Version: 0.0.1
#   LastChange: 2013-05-15 12:13:58
#      History:
=============================================================================*/
package net

import (
    . "github.com/sunminghong/letsgo/helper"
    //. "github.com/sunminghong/letsgo/log"
)

type LGMessageListWriter struct {
    *LGMessageWriter

    length int
    itemnum int

    meta []byte
}

func LGNewMessageListWriter(endian int) *LGMessageListWriter {
    list := &LGMessageListWriter{LGMessageWriter:LGNewMessageWriter(endian)}

    //LGTrace("messagelistwriter Init by called")
    list.init(768,endian)
    return list
}

func (list *LGMessageListWriter) WriteStartTag() {

    list.wind = 0
    list.maxInd = 0
}

func (list *LGMessageListWriter) WriteEndTag() {
    if list.itemnum ==0 {
        list.itemnum = list.maxInd
        list.needWriteMeta = false
    } else if list.maxInd != list.itemnum {
        panic("write list item num is wrong")
    }

    list.length ++

}

//对数据进行封包
func (list *LGMessageListWriter) ToBytes() []byte {
    list.metabuf.SetPos(0)
    list.buf.SetPos(0)
    //write heads
    heads,_ := list.metabuf.Read(4)

    //write list bytes length
    list.metabuf.Endianer.PutUint16(heads,
        uint16(list.buf.Len()+list.metabuf.Len() - 2))
    heads[2] = byte(list.length)

    //LGTrace("wind:",list.wind)
    heads[3] = byte(list.wind)
    //LGTrace("metabuflist",list.metabuf.Bytes())

    list.metabuf.Write(list.buf.Bytes())

    //LGTrace("metabuflist",list.metabuf.Bytes())
    return list.metabuf.Bytes()
}

type LGMessageListReader struct {
    *LGMessageReader

    //list length
    Length int

    //list byte length
    ByteLength int
}

func LGNewMessageListReader(buf *LGRWStream) *LGMessageListReader {
    list := &LGMessageListReader{LGMessageReader:&LGMessageReader{}}

    list.buf = buf


    byteLength,err := buf.ReadUint16()
    checkConvert(err)

    //n,data := buf.Read(int(byteLength))
    //if n!=int(byteLength) {
    //   checkConvert(ErrIndex)
    //}

    length,_ := buf.ReadByte()

    list.ByteLength = int(byteLength)
    list.Length = int(length)

    list.init()

    return list
}


func (list *LGMessageListReader) ReadStartTag() {
    if list.wind == 0 {
        return
    }

    //对齐列表项，如果列表数据项比读取的多，读下一个列表的数据是需要先将指针对齐
    for i:=list.wind;i<list.maxInd;i++ {
        ty,ok := list.meta[i]
        //LGTrace("checkread ty,ok",ty,ok)
        if !ok {
            continue
        }

        switch ty{
        case TY_UINT:
            list.buf.ReadUint()
        case TY_INT:
            list.buf.ReadInt()
        case TY_UINT16:
            list.buf.ReadUint16()
        case TY_UINT32:
            list.buf.ReadUint32()
        case TY_STRING:
            list.buf.ReadString()
        case TY_LIST:
            ll,_ := list.buf.ReadUint16()
            list.buf.Read(int(ll))
        }
    }
    list.wind = 0
}

func (list *LGMessageListReader) ReadEndTag() {
    list.wind = 0

    //对齐列表项，如果列表数据项比读取的多，读下一个列表的数据是需要先将指针对齐
}

