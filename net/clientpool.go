/*=============================================================================
#     FileName: clientpool.go
#         Desc: server base 
#       Author: sunminghong
#        Email: allen.fantasy@gmail.com
#     HomePage: http://weibo.com/5d13
#      Version: 0.0.1
#   LastChange: 2013-05-22 14:19:12
#      History:
=============================================================================*/
package net

import (
    "net"
    //"strconv"
    "time"
    "github.com/sunminghong/letsgo/log"
)

type ClientPool struct {
    read_buffer_size int

    newclient NewClientFunc
    datagram IDatagram
    /*runloop IRunLoop*/

    host string
    port int

    Clients       *ClientMap
    TransportNum int

    localhost string
    localport int

    Quit    chan bool
    boardcastChan    chan *DataPacket

    connaddr chan string
}

func NewClientPool(newclient NewClientFunc, datagram IDatagram /*,runloop IRunLoop*/) *ClientPool {
    cp := &ClientPool{Clients: NewClientMap()}
    cp.newclient = newclient
    cp.datagram = datagram
    /*
       if runloop != nil {
           c.runloop = runloop
       } else {
           c.runloop = NewRunLoop()
       }
    */

    cp.Quit = make(chan bool)
    cp.read_buffer_size = 1024


    //创建一个管道 chan map 需要make creates slices, maps, and channels only
    cp.boardcastChan = make(chan *DataPacket,1)
    go cp.boardcastHandler(cp.boardcastChan)

    return cp
}

func (cp *ClientPool) Start(name string,addr string,datagram IDatagram) {
    //go func() {
        ////Log("Hello Client!")

        //addr = host + ":" + strconv.Itoa(port)

        connection, err := net.Dial("tcp", addr)

        //mesg := "dialing"
        if err != nil {
            //Log("CLIENT: ERROR: ", mesg)
            return
        } else {
            //Log("Ok: ", mesg)
        }
        defer connection.Close()
        //Log("main(): connected ")

        newcid := cp.allocTransportid()

        if datagram == nil {
            datagram = cp.datagram
        }
        transport := NewTransport(newcid, connection, cp,datagram)
        client := cp.newclient(name,transport)
        cp.Clients.Add(newcid,name, client)

        //创建go的线程 使用Goroutine
        go cp.transportSender(transport)
        go cp.transportReader(transport, client)


        time.Sleep(2)

        <-transport.Quit
    //}()
    <-cp.Quit
}

func (cp *ClientPool) SetMaxConnections(max int) {

}

func (cp *ClientPool) Close(cid int) {
    if cid == 0 {
        for _, client := range cp.Clients.All(){
            //c.running[cid] = false
            client.GetTransport().Quit <- true
        }
        return
    }

    //c.running[cid] = false
    cp.Clients.Get(cid).GetTransport().Quit <- true
}

func (cp *ClientPool) removeClient(cid int) {
    cp.Clients.Remove(cid)
}

func (cp *ClientPool) allocTransportid() int {
    cp.TransportNum += 1
    return cp.TransportNum
}

func (cp *ClientPool) transportReader(transport *Transport, client IClient) {
    buffer := make([]byte, cp.read_buffer_size)
    for {

        bytesRead, err := transport.Conn.Read(buffer)

        if err != nil {
            client.Closed()
            transport.Closed()
            cp.removeClient(transport.Cid)
            //Log(err)
            break
        }

        //Log("read to buff:", bytesRead)
        transport.BuffAppend(buffer[0:bytesRead])

        //Log("transport.Buff", transport.Stream.Bytes())
        n, dps := cp.datagram.Fetch(transport)
        //Log("fetch message number", n)
        if n > 0 {
            client.ProcessDPs(dps)
        }
    }
    //Log("TransportReader stopped for ", transport.Cid)
}

func (cp *ClientPool) transportSender(transport *Transport) {
    for {
        select {
        case dp := <-transport.outgoing:
            log.Trace("clientpool transportSender:",dp.Type, dp.Data)
            //buf := cp.datagram.Pack(dp)
            //transport.Conn.Write(buf)

            cp.datagram.PackWrite(transport.Conn.Write,dp)
        case <-transport.Quit:
            //Log("Transport ", transport.Cid, " quitting")
            transport.Conn.Close()

            //client.Closed()
            transport.Closed()
            cp.removeClient(transport.Cid)
            break
        }
    }
}

func (cp *ClientPool) boardcastHandler(boardcastChan <-chan *DataPacket) {
    for {
        //在go里面没有while do ，for可以无限循环
        //Log("boardcastHandler: chan Waiting for input")
        dp := <-boardcastChan
        //buf := c.datagram.pack(dp)

        sendCid := dp.FromCid
        for Cid, c := range cp.Clients.All() {
            if sendCid == Cid {
                continue
            }
            c.GetTransport().outgoing <- dp
        }
        //Log("boardcastHandler: Handle end!")
    }
}

//send boardcast message data for other object
func (cp *ClientPool) SendBoardcast(transport *Transport, dp *DataPacket) {
    cp.boardcastChan <- dp
}
