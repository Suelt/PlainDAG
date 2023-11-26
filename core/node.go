package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/PlainDAG/go-PlainDAG/config"
	"github.com/PlainDAG/go-PlainDAG/core/ttype"

	"github.com/PlainDAG/go-PlainDAG/sign"
)

type Node struct {
	//DAG ledger structure
	bc *Blockchain `json:"bc"`

	handler Messagehandler
	//thread-safe integer
	currentround atomic.Int64 `json:"currentround"`
	trans        *NetworkTransport
	cfg          *config.Config

	ls        *LeaderSelector
	committer *StaticCommitter
	txPool    *TxPool

	start       time.Time
	latestRound int
	staticTx    *ttype.BatchTx
	syncNumber  atomic.Int32
	syncChan    chan bool
}

func (n *Node) genTrans(rn int) (Message, error) {
	if rn < 3 {
		return n.genBasicMsg(rn)
	}
	if rn%100 == 0 {
		log.Println("generate transaction for round" + strconv.Itoa(rn))
	} //generate transaction
	if rn%rPerwave == 0 {

		return n.genFroundMsg(rn)
	} else if rn%rPerwave == 1 {
		return n.genBasicMsg(rn)
	} else if rn%rPerwave == 2 {
		return n.genLroundMsg(rn)
	}
	return nil, nil
}

func (n *Node) genFroundMsg(rn int) (*FRoundMsg, error) {

	basic, err := n.genBasicMsg(rn)
	// lastRound := n.bc.GetRound(rn - 1)
	// indexes, error := lastRound.getIndexByRefsBatch(basic.GetRefs())
	// //fmt.Println(len(lm.ARefs))
	// if error != nil {
	// 	return nil, error

	// }
	//fmt.Println("generating f-round message at round ", basic.GetRN(), " and it References ", indexes)
	if err != nil {
		return nil, err
	}
	if rn == 3 {
		embededrefs := make(map[int][][]byte)
		refs := basic.GetRefs()
		embededrefs[2] = refs

		secondRound := n.bc.GetRound(2)
		msgs, err := secondRound.getMsgByRefsBatch(refs)
		if err != nil {
			return nil, err
		}
		// go func() {
		// 	for _, msg := range msgs {
		// 		n.committer.removeUnembeded(2, n.cfg.StringIdMap[string(msg.GetSource())])
		// 	}
		// }()

		refs, err = getRefsByMsgsBatch(msgs)

		if err != nil {
			return nil, err
		}
		embededrefs[1] = refs

		//firstRound := n.bc.GetRound(1)
		//msgs, err = firstRound.getMsgByRefsBatch(refs)
		if err != nil {
			return nil, err
		}
		// var wg sync.WaitGroup
		// wg.Add(1)
		// go func() {
		// 	for _, msg := range msgs {
		// 		n.committer.removeUnembeded(1, n.cfg.StringIdMap[string(msg.GetSource())])
		// 	}
		// 	wg.Done()
		// }()
		// wg.Wait()
		return NewFRoundMsg(basic, 0, nil, 1, embededrefs), nil

	} else {
		// t := time.Now().UnixMilli()
		embededLeaderRound, err := n.findLastEmbededLeader(basic)
		if err != nil {
			return nil, err
		}
		//fmt.Println("node "+strconv.Itoa(n.cfg.Id)+"  is Generating f-round message at round   "+strconv.Itoa(basic.GetRN())+",   embeded leader is  at round", embededLeaderRound, "and is at wave ", embededLeaderRound/3)
		// if embededLeaderRound != 0 {
		// 	//fmt.Println("the embeded leaderslot is ", *n.ls.slotMap[embededLeaderRound/3])
		// }
		embededrefs, err, LeastEmbededRn := n.embedMessages(basic, embededLeaderRound)
		if err != nil {
			return nil, err
		}
		// t2 := time.Now().UnixMilli()
		// fmt.Println(t2 - t)
		// for i := LeastEmbededRn; i < rn; i++ {
		// 	round := n.bc.GetRound(i)
		// 	msgs, _ := round.getMsgByRefsBatch(embededrefs[i])
		// 	for _, msg := range msgs {
		// 		//fmt.Println("embed message at round", i, "from node ", n.cfg.StringIdMap[string(msg.GetSource())])
		// 	}

		// }

		return NewFRoundMsg(basic, embededLeaderRound, nil, LeastEmbededRn, embededrefs), nil
	}

}

func (n *Node) findLastEmbededLeader(msg Message) (int, error) {
	traceBackWave := 1
	middleRounds := make([]*Round, 0)
	for {
		n.ls.slotMapLock.Lock()
		slot := n.ls.slotMap[msg.GetRN()/3-traceBackWave]
		n.ls.slotMapLock.Unlock()

		roundForSlot := n.bc.GetRound(msg.GetRN() - traceBackWave*3)

		for i := msg.GetRN() - (traceBackWave-1)*3 - 1; i > msg.GetRN()-traceBackWave*3; i-- {
			middleRounds = append(middleRounds, n.bc.GetRound(i))
		}

		// for _, Round := range middleRounds {
		// 	fmt.Println("round ", Round.roundNumber)
		// }

		//fmt.Println("what?")
		roundForSlot.messageLock.Lock()

		//fmt.Println(*slot)

		if len(roundForSlot.msgs[*slot]) == 0 {
			roundForSlot.messageLock.Unlock()
			//fmt.Println("nil slot, now searching for lower wave")
			traceBackWave++
			if msg.GetRN()/3-traceBackWave <= 0 {
				return 0, nil
			} else {
				middleRounds = append(middleRounds, roundForSlot)
				continue
			}
		}

		msgs := roundForSlot.msgs[*slot]
		roundForSlot.messageLock.Unlock()
		for _, m := range msgs {
			//fmt.Println("here?")
			havepath, err := msg.HavePath(m, middleRounds, roundForSlot)
			if err != nil {

				return 0, err
			}
			if havepath {

				return msg.GetRN() - traceBackWave*3, nil
			} else {
				traceBackWave++
				if msg.GetRN()/3-traceBackWave <= 0 {
					return 0, nil
				} else {
					middleRounds = append(middleRounds, roundForSlot)
					continue
				}
			}
		}

	}
}

func (n *Node) embedMessages(NewGen Message, embededLeaderRound int) (map[int][][]byte, error, int) {

	currentRound := NewGen.GetRN()
	Refs := NewGen.GetRefs()
	embededrefs := make(map[int][][]byte)

	for i := currentRound - 1; i >= embededLeaderRound && i > 0; i-- {
		round := n.bc.GetRound(i)
		msgs, err := round.getMsgByRefsBatch(Refs)
		// for _, msg := range msgs {
		// 	n.committer.removeUnembeded(i, n.cfg.StringIdMap[string(msg.GetSource())])
		// }
		if err != nil {
			return nil, err, 0
		}
		embededrefs[i] = Refs
		Refs, err = getRefsByMsgsBatch(msgs)
		if err != nil {
			return nil, err, 0
		}

	}

	//delete leaderslot hash in embededrefs[embededLeaderslot]
	if embededLeaderRound == 0 {

		return embededrefs, nil, 1
	} else {
		leaderSlot := n.ls.slotMap[embededLeaderRound/3]

		lastLeaderRound := n.bc.GetRound(embededLeaderRound)
		lastLeaderRound.messageLock.Lock()
		lastLeaderMsgs := lastLeaderRound.msgs[*leaderSlot]
		lastLeaderRound.messageLock.Unlock()

		hash := lastLeaderMsgs[0].GetHash()
		for i := 0; i < len(embededrefs[embededLeaderRound]); i++ {
			if bytes.Equal(embededrefs[embededLeaderRound][i], hash) {
				embededrefs[embededLeaderRound] = append(embededrefs[embededLeaderRound][:i], embededrefs[embededLeaderRound][i+1:]...)
				break
			}
		}

		lastLeaderRefs := lastLeaderMsgs[0].GetRefs()
		sort.Slice(lastLeaderRefs, func(i, j int) bool {
			return bytes.Compare(lastLeaderRefs[i], lastLeaderRefs[j]) < 0
		})
		sort.Slice(Refs, func(i, j int) bool {
			return bytes.Compare(Refs[i], Refs[j]) < 0
		})
		var least int
		least = embededLeaderRound
		if len(lastLeaderRefs) == len(Refs) {
			for i := 0; i < len(lastLeaderRefs); i++ {
				if !bytes.Equal(lastLeaderRefs[i], Refs[i]) {
					earlierRefs, leastRn, _ := n.searchForEarlierEmbeded(Refs, embededLeaderRound)
					least = leastRn
					for k, v := range earlierRefs {
						embededrefs[k] = v
					}
					break
				}
			}

		} else {
			earlierRefs, leastRn, _ := n.searchForEarlierEmbeded(Refs, embededLeaderRound)
			least = leastRn
			for k, v := range earlierRefs {
				embededrefs[k] = v
			}
		}

		return embededrefs, nil, least
	}

}

func (n *Node) searchForEarlierEmbeded(refs [][]byte, roundNumber int) (map[int][][]byte, int, error) {
	lastEmbededRoundNumber, embededRefs := n.retLastEmbededRoundNumber(roundNumber)
	//fmt.Println("last embeded round number is ", lastEmbededRoundNumber, "all the former round does not need to be embeded")
	retMap := make(map[int][][]byte)
	for i := roundNumber - 1; i > lastEmbededRoundNumber; i-- {
		unique := [][]byte{}
		m := make(map[string]bool)
		for _, b := range embededRefs[i] {
			m[string(b)] = true
		}
		for _, b := range refs {
			if !m[string(b)] {
				// If the byte is not in the map, it's unique
				unique = append(unique, b)
			}
		}
		retMap[i] = unique
		round := n.bc.GetRound(i)
		msgs, err := round.getMsgByRefsBatch(unique)
		if err != nil {
			return nil, 0, err
		}
		refs, err = getRefsByMsgsBatch(msgs)
		if err != nil {
			return nil, 0, err
		}

	}
	return retMap, lastEmbededRoundNumber, nil
}

func (n *Node) retLastEmbededRoundNumber(lastLeaderRoundNumber int) (int, map[int][][]byte) {
	if lastLeaderRoundNumber == 0 {
		return 0, nil
	}
	leaderSlot := n.ls.slotMap[lastLeaderRoundNumber/3]
	lastLeaderRound := n.bc.GetRound(lastLeaderRoundNumber)
	lastLeaderRound.messageLock.Lock()
	lastLeaderMsgs := lastLeaderRound.msgs[*leaderSlot]
	lastLeaderRound.messageLock.Unlock()
	traceBackRound := 1
	refs := lastLeaderMsgs[0].GetRefs()
	embededRefs := make(map[int][][]byte)

	for {
		embededRefs[lastLeaderRoundNumber-traceBackRound] = refs
		if len(refs) == 5*f+1-n.cfg.Byzantine {
			return lastLeaderRoundNumber - traceBackRound, embededRefs

		}
		round := n.bc.GetRound(lastLeaderRoundNumber - traceBackRound)
		msgs, err := round.getMsgByRefsBatch(refs)
		if err != nil {
			return 0, nil
		}
		refs, err = getRefsByMsgsBatch(msgs)
		if err != nil {
			return 0, nil
		}
		traceBackRound++
		if lastLeaderRoundNumber-traceBackRound < 1 {
			return 0, embededRefs
		}
	}

}
func (n *Node) genLroundMsg(rn int) (*LRoundMsg, error) {
	basic, err := n.genBasicMsg(rn)
	if err != nil {
		return nil, err
	}

	lround := n.bc.GetRound(rn - 2)
	mround := n.bc.GetRound(rn - 1)
	bytes, err := lround.genArefs(basic, mround)
	if err != nil {
		return nil, err
	}

	lroundmsg, err := NewLroundMsg(bytes, basic)

	if err != nil {
		return nil, err
	}
	// indexes, err := lround.getIndexByRefsBatch(bytes)
	// if err != nil {
	// 	return nil, err
	// }
	//fmt.Println("Generated lround message at round ", rn, "and A-references ", indexes)
	return lroundmsg, nil

}

func (n *Node) genBasicMsg(rn int) (*BasicMsg, error) {
	lastRound := n.bc.GetRound(rn - 1)
	if lastRound == nil {
		return nil, errors.New("last round is nil")
	}
	//generate transaction
	refsByte := lastRound.retMsgsToRef()
	//TODO: adjust block concurrency

	// go n.stateDB.BatchCreateObjects(txs)
	// <-n.txPool.hasblock
	// fmt.Println("generate transaction for round" + strconv.Itoa(rn))
	txs := n.txPool.ExtractTxs()
	if txs == nil {
		// fmt.Println("nil batch for round", rn)
		txs = &ttype.BatchTx{
			Timestamp: time.Now().UnixMilli(),
		}
	}
	// fmt.Println(len(txs.Txs))
	basicMsg, err := NewBasicMsg(rn, refsByte, n.cfg.IdPubkeymap[n.cfg.Id], *txs)
	if err != nil {
		return nil, err
	}
	//fmt.Println("ends here?")
	return basicMsg, nil

}

func (n *Node) genThresMsg(rn int) *ThresSigMsg {

	bytes := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(bytes, int64(rn))
	//fmt.Println("generated for round  ", rn, "    ", bytes)
	s := sign.SignTSPartial(n.cfg.TSPrvKey, bytes)
	thresSigMsg := &ThresSigMsg{
		Wn:     rn / rPerwave,
		Sig:    s,
		Source: n.cfg.Pubkeyraw,
	}
	//fmt.Println(thresSigMsg.source)
	return thresSigMsg

}
func (n *Node) paceToNextRound() (Message, error) {
	//generate transaction
	rn := int(n.currentround.Load())
	n.handler.buildContextForRound(rn + 1)
	//this removal is only used to save memory when the code is not finished.
	// if rn > 11 {
	// 	minustenRound := n.bc.GetRound(rn - 10)
	// 	minustenRound.rmvAllMsgsWhenCommitted()
	// }
	//fmt.Println("generating?")
	// t1 := time.Now().UnixMilli()
	msg, err := n.genTrans(rn + 1)
	if err != nil {
		return nil, err
	}
	// t2 := time.Now().UnixMilli()
	// fmt.Println("timepast when generating a message: ", t2-t1, "roundnumber: ", msg.GetRN())
	// fmt.Println("generated?")

	n.committer.addToUnCommitted(msg, n.cfg.Id)
	//n.committer.addToUnEmbeded(msg, n.cfg.Id)
	// if msg.GetRN()%3 != 2 || msg.GetRN() == 2 {
	//lround := n.bc.GetRound(msg.GetRN() - 1)
	// if (rn+1)%3 != 2 || (rn+1) == 2 {
	// 	indexes, error := lround.getIndexByRefsBatch(msg.GetRefs())
	// 	//fmt.Println(len(lm.ARefs))
	// 	if error != nil {
	// 		panic(error)
	// 	}

	// 	//fmt.Println("received message at round ", msg.GetRN(), "from node", n.cfg.StringIdMap[string(msg.GetSource())], " and it References ", indexes)
	// 	// } //fmt.Println("ends here tryhandle2?")
	// }
	// msgbytes, sig, err := utils.MarshalAndSign(msg, n.cfg.Prvkey)
	// if err != nil {
	// 	return nil, err
	// }

	if rn < 2 {
		//fmt.Println("sent?")
		go n.SendMsgToAll(1, msg, []byte{})
	} else {
		msgtype := (rn + 1) % 3
		//fmt.Println(msgtype)
		//fmt.Println((rn + 1) % 3)
		go n.SendMsgToAll(uint8(msgtype), msg, []byte{})
	}
	if rn%rPerwave == 1 && rn != 1 {
		thresSigMsg := n.genThresMsg(rn + 1)

		//fmt.Println(thresSigMsg.sig, thresSigMsg.source, thresSigMsg.wn)
		// thresSigMsgBytes, sig, err := utils.MarshalAndSign(thresSigMsg, n.cfg.Prvkey)
		n.handler.handleThresMsg(thresSigMsg, []byte{})
		if err != nil {
			return nil, err
		}
		//rintln(thresSigMsgBytes)
		go n.SendMsgToAll(3, thresSigMsg, []byte{})
	}

	//initialize a new round with the newly generated message msg
	newR, err := newRound(rn+1, msg, n.cfg.Id)

	if err != nil {
		return nil, err
	}
	n.bc.AddRound(newR)
	msg.AfterAttach(n)
	n.currentround.Add(1)

	go n.handler.handleFutureVers(rn + 1)
	//n.SendMsgToAll(1, msgbytes, sig)
	return msg, err
}

func (n *Node) HandleMsgForever() {
	if n.cfg.Nodetype == 1 {
		return
	}
	rpcCh := n.trans.Consumer()
	for {
		select {

		case msg := <-rpcCh:
			// log.Println("receive msg: ", reflect.TypeOf(msg.Msg))
			switch msgAsserted := msg.Command.(type) {
			case Message:

				go n.handleMsg(msgAsserted, msg.Sig)
			case *ThresSigMsg:
				// fmt.Println("received thresmsg")
				//fmt.Println(msg.Msg)
				go n.handleThresMsg(msgAsserted, msg.Sig)
				// case *CShardMsg:
				// 	// log.Println("[HandleMsgForever] receive a cross-shard msg, ", msgAsserted)
				// 	go n.handleCShardMsg(msgAsserted, msg.Sig, msg.Msgbytes)
				// }
			case *SyncMsg:
				go n.handleSyncMsg(msgAsserted, msg.Sig)
			}

		}
	}
}

func (n *Node) handleSyncMsg(msg *SyncMsg, sig []byte) {
	n.syncNumber.Add(1)
	if n.syncNumber.Load() == int32(5*f-n.cfg.Byzantine) {
		n.syncChan <- true
	}
}

// func (n *Node) HandleTxForever() {
// 	log.Println("HandleTxForever...")
// 	for {
// 		select {
// 		case t := <-n.network.ExtractTx():
// 			switch msgAsserted := t.Msg.(type) {
// 			case *ttype.Transaction:
// 				// log.Println("receive tx", msgAsserted.Payload)
// 				go n.handleTx(msgAsserted)
// 			}
// 		}
// 	}
// }

func (n *Node) handleMsg(msg Message, sig []byte) {
	// t1 := time.Now().UnixMilli()
	if err := n.handler.handleMsg(msg, sig); err != nil {
		panic(err)
	}
	// t2 := time.Now().UnixMilli()
	// fmt.Println("timepast when handling a message: ", t2-t1)
	//go func() {
	//	var txs []*ttype.Transaction
	//	plains := msg.GetPlainMsgs()
	//	for _, plain := range plains {
	//		if len(plain) > 0 {
	//			var t ttype.Transaction
	//			if err := json.Unmarshal(plain, &t); err != nil {
	//				log.Panic(err)
	//			}
	//			txs = append(txs, &t)
	//		}
	//	}
	//	n.stateDB.BatchCreateObjects(txs)
	//}()
}

func (n *Node) handleThresMsg(msg *ThresSigMsg, sig []byte) {
	if err := n.handler.handleThresMsg(msg, sig); err != nil {
		panic(err)
	}
}

//	func (n *Node) handleCShardMsg(msg *CShardMsg, sig []byte, msgbytes []byte) {
//		if err := n.handler.handleCShardMsg(msg, sig, msgbytes); err != nil {
//			panic(err)
//		}
//		// log.Println("[handleCShardMsg] done")
//	}
// func (n *Node) handleTx(tx *ttype.Transaction) {
// 	n.txPool.pending.Append(tx)
// }

func (n *Node) SendMsgToAll(messagetype uint8, msg interface{}, sig []byte) error {

	for i, addr := range n.cfg.IdaddrMap {

		go func(index uint32, addr string) {
			var conn *NetConn
			var err error
			conn, _ = n.trans.GetConn(addr+":"+strconv.Itoa(n.cfg.IdportMap[int(index)]), 0, index)

			//  defer n.trans.ReturnConn(conn)
			if int(index) == n.cfg.Id {
				// fmt.Println("did you finish===")
				time.Sleep(time.Duration(10) * time.Millisecond)
			} else {

				err = SendRPC(conn, MsgType(messagetype), msg, sig)
				if err != nil {
					fmt.Println("cannot send rpc:", err)
					//return err
				}
			}

			if err = n.trans.ReturnConn(conn); err != nil {
				fmt.Println("cannot release conn:", err)
			}
		}(uint32(i), addr)

	}
	return nil
}

func (n *Node) SendForever() {
	if n.cfg.Nodetype == 1 {
		return
	}

	msg := &SyncMsg{
		id: n.cfg.Id,
	}
	go n.SendMsgToAll(4, msg, []byte{})

	<-n.syncChan

	go n.txPool.GenerateTxs(n.cfg.Batchtimeout, n.cfg.Batchsize)
	for {

		n.handler.readyForRound(int(n.currentround.Load()) + 1)
		_, err := n.paceToNextRound()
		if err != nil {
			panic(err)
		}

		//msg.DisplayinJson()

		//time.Sleep(100 * time.Millisecond)
		// H := []byte{1, 2, 3}

		// refs := make([][]byte, 0)
		// refs = append(refs, H)

		// msg, err := NewMroundmsg(1, refs, n.cfg.Pubkeyraw)
		// if err != nil {
		// 	panic(err)
		// }
		// // for _, peer := range n.network.H.Peerstore().Peers() {
		// // 	s := peer.Pretty()

		// // 	fmt.Println(s)
		// // }
		// msgbytes, err := json.Marshal(msg)
		// if err != nil {
		// 	panic(err)
		// }

		// sig, err := n.cfg.Prvkey.Sign(msgbytes)
		// if err != nil {
		// 	panic(err)
		// }
		// err = n.SendMsgToAll(2, msg, sig)
		// if err != nil {
		// 	panic(err)
		// }
	}

}

func (n *Node) serialize(filepath string) error {
	//Todo
	//serialize the committed messages to the database
	return nil
}

// // prefetchStates fetches the states of accounts into the statedb
// func (n *Node) prefetchStates(msgs []Message) {
// 	for _, msg := range msgs {
// 		plains := msg.GetPlainMsgs()
// 		for _, plain := range plains {
// 			if len(plain) > 0 {
// 				var t ttype.Transaction
// 				if err := json.Unmarshal(plain, &t); err != nil {
// 					log.Panic(err)
// 				}
// 				payload := tx.Data()
// 				for addr := range payload {
// 					n.stateDB.PreFetch(addr)
// 				}
// 			}
// 		}
// 	}
// }

// executeTxs processes transactions in all committed messages
// func (n *Node) processTxs(msgs []Message) {
// 	n.executeLock.Lock()
// 	defer n.executeLock.Unlock()
// 	start := time.Now()
// 	load, output, txs := n.rmDuplicatedTxs(msgs)
// 	go func() {
// 		// persist msg storage
// 		StoreMsgs(n.database, msgs)
// 		// persist tx storage
// 		for _, tt := range txs {
// 			StoreTxs(n.database, tt)
// 		}
// 	}()
// 	executor := NewExecutor(n.stateDB, MaximumProcessors, load, n)
// 	stateRoot := executor.Processing(txs, output, Frequency)
// 	duration := time.Since(start)
// 	log.Printf("Time of processing transactions is: %s, the generated state root is: %v\n", duration, stateRoot)
// }

// rmDuplicatedTxs deletes duplicate transactions for transaction execution
// func (n *Node) rmDuplicatedTxs(msgs []Message) (int, []string, map[string][]*ttype.Transaction) {
// 	var load int
// 	var output []string
// 	var deleted = make(map[string]struct{})
// 	var txs = make(map[string][]*ttype.Transaction)
// 	var processing = make(map[string][]*ttype.Transaction)

// 	// obtain messages and convert to transactions
// 	for _, msg := range msgs {
// 		id := string(msg.GetHash())
// 		output = append(output, id)

// 		plains := msg.GetPlainMsgs()
// 		for _, plain := range plains {
// 			if len(plain) > 0 {
// 				var tx ttype.Transaction
// 				if err := json.Unmarshal(plain, &tx); err != nil {
// 					log.Panic(err)
// 				}
// 				txs[id] = append(txs[id], &tx)
// 			}
// 		}
// 	}

// 	for _, id := range output {
// 		var added []*ttype.Transaction
// 		blk := txs[id]
// 		// delete committed txs from tx pool
// 		n.txPool.DeleteTxs(blk)
// 		for _, t := range blk {
// 			iid := string(t.ID)
// 			if _, ok := deleted[iid]; !ok {
// 				added = append(added, t)
// 				deleted[iid] = struct{}{}
// 			}
// 		}
// 		processing[id] = added
// 		load += len(added)
// 	}

// 	fmt.Printf("tx load is: %d\n", load)

// 	return load, output, processing
// }

//func StartandConnect() (*Node, error) {
//	index := flag.Int("f", 0, "config file path")
//	flag.Parse()
//	//convert int to string
//	filepath := "node" + strconv.Itoa(*index)
//	n, err := NewNode(filepath)
//	if err != nil {
//		return nil, err
//	}
//
//	time.Sleep(15 * time.Second)
//	err = n.connecttoOthers()
//	if err != nil {
//		return nil, err
//	}
//	n.constructpubkeyMap()
//	// get the pubkey of my own host
//
//	return n, nil
//}

// func (n *Node) SendTxsForLoop(cycle, rate int) {
// 	fmt.Println("Sending txs...")
// 	// 250bytes
// 	messageconst = []byte{154, 80, 2, 82, 229, 242, 220, 255, 179, 87, 61, 154, 8, 88, 8, 107, 15, 130, 189, 156, 210, 66, 119, 158, 22, 164, 100, 166, 125, 222, 189, 140, 149, 138, 224, 105, 95, 112, 255, 126, 180, 47, 154, 161, 172, 224, 168, 68, 205, 1, 82, 65, 119, 220, 239, 199, 105, 36, 211, 130, 17, 17, 103, 221, 81, 251, 40, 174, 56, 32, 146, 64, 32, 181, 80, 209, 211, 86, 83, 153, 68, 131, 145, 200, 112, 162, 8, 165, 245, 11, 186, 213, 79, 2, 56, 69, 144, 62, 66, 63, 226, 226, 183, 23, 230, 176, 191, 150, 200, 66, 1, 221, 85, 140, 19, 251, 66, 183, 61, 235, 12, 47, 212, 153, 66, 125, 132, 214, 184, 218, 185, 125, 118, 61, 102, 15, 180, 44, 230, 134, 105, 13, 127, 44, 250, 1, 224, 47, 241, 108, 120, 95, 125, 49, 191, 125, 135, 222, 211, 120, 82, 31, 103, 199, 193, 217, 50, 34, 78, 214, 131, 99, 95, 18, 235, 235, 180, 40, 33, 188, 178, 39, 143, 147, 167, 96, 78, 150, 248, 165, 91, 10, 138, 102, 214, 206, 176, 200, 85, 185, 53, 121, 76, 116, 151, 119, 155, 76, 16, 211, 193, 184, 250, 202, 83, 91, 147, 87, 31, 234, 191, 1, 114, 192, 255, 105, 110, 14, 98, 57, 110, 87, 100, 154, 188, 38, 84, 87, 137, 200, 60, 51, 84, 216, 201, 82, 11, 170, 16, 52}
// 	consttx := ttype.Transaction{Constbytes: messageconst}

// 	txdata, _ := json.Marshal(consttx)

// 	for i := 0; i < cycle; i++ {
// 		for j := 0; j < rate; j++ {

// 			if err := n.network.Broadcast(TxTag, txdata, []byte("tx")); err != nil {
// 				log.Println(err)
// 			}
// 		}
// 	}
// }

// func (n *Node) sendTxsForLoop(cycle, rate int) {
// 	log.Println("sendTxsForLoop...")
// 	for i := 0; i < cycle; i++ {
// 		r := rand.New(rand.NewSource(time.Now().UnixNano()))
// 		z := zipf.NewZipf(r, Skew, 100000)
// 		for j := 0; j < rate; j++ {
// 			t := CreateMimicWorkload(Ratio, z, 2)
// 			n.txPool.pending.Append(t)
// 			txdata, _ := json.Marshal(*t)
// 			if err := n.SendMsgToAll(TxTag, txdata, []byte("tx")); err != nil {
// 				log.Println(err)
// 			}
// 		}
// 	}
// }

func (n *Node) StartListen() error {

	addr := ":" + strconv.Itoa(n.cfg.Port)

	var err error
	n.trans, err = NewTCPTransport(addr, 10*time.Second, nil, 10000)
	if err != nil {
		return err
	}

	return nil

}

func (n *Node) StartConnect() error {
	var err error
	if err = n.establishConns(); err != nil {
		//return err
	}

	return nil
}

func (n *Node) establishConns() error {
	if n.trans == nil {
		return errors.New("networktransport has not been created")
	}

	for i, addr := range n.cfg.IdaddrMap {
		// Avoid establishing connection with itself
		if i == n.cfg.Id {
			continue
		}
		addr = addr + ":" + strconv.Itoa(n.cfg.IdportMap[i])

		conn, err := n.trans.GetConn(addr, 0, 0)
		if err != nil {
			fmt.Println(err)
			continue
			//return err
		}

		n.trans.ReturnConn(conn)
		fmt.Printf("Connection from %s to %s has been established\n", n.cfg.Ipaddress+":"+strconv.Itoa(n.cfg.Port), addr)
	}

	return nil
}

func StartandConnect(config_path string) (*Node, error) {
	n, err := NewNode(config_path)
	if err != nil {
		return nil, err
	}

	if err = n.StartListen(); err != nil {
		panic(err)
	}

	// wait for each node to start
	time.Sleep(time.Second * 90)

	if err = n.StartConnect(); err != nil {
		panic(err)
	}

	// n.constructpubkeyMap()
	// go n.txPool.GenerateTxs(n.cfg.Batchtimeout, n.cfg.Batchsize)
	return n, nil
}

// // For test
// func StartandConnect() (*Node, error) {
// 	index := flag.Int("f", 0, "config file path")
// 	flag.Parse()
// 	//convert int to string
// 	filepath := "node" + strconv.Itoa(*index)
// 	n, err := NewNode(filepath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	time.Sleep(20 * time.Second)
// 	err = n.connecttoOthers()
// 	if err != nil {
// 		return nil, err
// 	}
// 	n.constructpubkeyMap()

// 	return n, nil

// }

/*
This function is used to test the sending and receiving of cross-shard messages
*/
