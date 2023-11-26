package p2p

import (
	"log"
	"reflect"
	"sync"

	"github.com/hashicorp/go-msgpack/codec"
	"github.com/libp2p/go-libp2p/core/crypto"
)

func Startpeer(port int, prvkey crypto.PrivKey, reflectedTypesMap map[uint8]reflect.Type) (*NetworkDealer, error) {
	n, err := NewnetworkDealer(port, prvkey, reflectedTypesMap)
	if err != nil {
		return nil, err
	}
	n.Listen()
	return n, nil

}

func (n *NetworkDealer) Connectpeers(peerid int, idaddrmap map[int]string, idportmap map[int]int, pubstringsmap map[int]string) error {
	log.Println("connect to ", len(idaddrmap)-1, "peers in local shard")
	for id, addr := range idaddrmap {
		if id != peerid {

			writer, err := n.Connect(idportmap[id], addr, pubstringsmap[id])
			if err != nil {
				return err
			}
			log.Println("connect to ", addr, idportmap[id], " success")
			destMultiAddr := PackMultiaddr(idportmap[id], addr, pubstringsmap[id])
			log.Println("dest: ", destMultiAddr)
			n.connPool[destMultiAddr] = &conn{
				w: writer,

				dest: destMultiAddr,

				encode: codec.NewEncoder(writer, &codec.MsgpackHandle{}),
			}
		}
	}
	return nil

}

// func (n *NetworkDealer) PrintConnPool() {
// 	for k, v := range n.connPool {
// 		log.Println(k, v)
// 	}

// }

func (n *NetworkDealer) Broadcast(messagetype uint8, msg interface{}, sig []byte) error {
	// log.Println("Broadcast msg type: ", messagetype)

	n.BroadcastSyncLock.Lock()

	//time.Sleep(time.Duration(n.latencyrand.Poisson(simlatency)) * time.Millisecond)
	var wg sync.WaitGroup
	// fmt.Println("round 2?")
	for _, conn := range n.connPool {
		//fmt.Println(conn)
		wg.Add(1)
		c := conn
		go func() {

			err := n.SendMsg(messagetype, msg, sig, c.dest, false)
			if err != nil {
				// reconnect the node and resend
				success := false
				for i := 2; i > 0; i-- {
					log.Println("fail to send msg to dest: ", c.dest, "try reconnecting and sending again")
					err := n.SendMsg(messagetype, msg, sig, c.dest, true)
					if err == nil {
						success = true
						break
					}
				}
				if !success {
					log.Println("fail to send msg to dest after retry: ", c.dest, "try reconnecting and sending again")
				}
			}
			wg.Done()

		}()
	}
	wg.Wait()
	//fmt.Println("are you finished?")
	n.BroadcastSyncLock.Unlock()
	return nil

}

/*
broad a msg to shard shardid
@param destList a list of destMultiAddr like "/ip4/0.0.0.0/tcp/9000/p2p/QmUSiw4JrFmqf4mAPozaSas2mAY8vxBow3gzyZtXdPPEqV"
*/

// func (n *NetworkDealer) HandleMsgForever() {

// 	for {
// 		select {
// 		case <-n.shutdownCh:
// 			return
// 		case msg := <-n.msgch:
// 			log.Println("receive msg: ", msg.Msg)
// 		}

// 	}
// }

// func printconfig(c *config.P2pconfig) {
// 	log.Println("id: ", c.Id)
// 	log.Println("nodename: ", c.Nodename)
// 	log.Println("port: ", c.Port)
// 	log.Println("addr: ", c.Ipaddress)
// 	log.Println("idportmap: ", c.IdportMap)
// 	log.Println("idaddrmap: ", c.IdaddrMap)
// 	log.Println("pubkeyothersmap: ", c.Pubkeyothersmap)
// 	log.Println("prvkey: ", c.Prvkey)
// 	log.Println("pubkey: ", c.Pubkey)

// }
