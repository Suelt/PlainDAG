package core

import (
	"sync"
	"time"

	"github.com/PlainDAG/go-PlainDAG/core/ttype"
)

type TxPool struct {
	pending  *pendingPool
	n        int
	hasblock chan bool
}

func NewTxPool() *TxPool {
	pool := &TxPool{
		pending:  newPending(),
		hasblock: make(chan bool, 10000),
	}

	return pool
}

func (tp *TxPool) GenerateTxs(batchtimeout int, batchsize int) {
	for {
		time.Sleep(time.Duration(batchtimeout) * time.Millisecond)
		messageconst := []byte{154, 80, 2, 82, 229, 242, 220, 255, 179, 87, 61, 154, 8, 88, 8, 107, 15, 130, 189, 156, 210, 66, 119, 158, 22, 164, 100, 166, 125, 222, 189, 140, 149, 138, 224, 105, 95, 112, 255, 126, 180, 47, 154, 161, 172, 224, 168, 68, 205, 1, 82, 65, 119, 220, 239, 199, 105, 36, 211, 130, 17, 17, 103, 221, 81, 251, 40, 174, 56, 32, 146, 64, 32, 181, 80, 209, 211, 86, 83, 153, 68, 131, 145, 200, 112, 162, 8, 165, 245, 11, 186, 213, 79, 2, 56, 69, 144, 62, 66, 63, 226, 226, 183, 23, 230, 176, 191, 150, 200, 66, 1, 221, 85, 140, 19, 251, 66, 183, 61, 235, 12, 47, 212, 153, 66, 125, 132, 214, 184, 218, 185, 125, 118, 61, 102, 15, 180, 44, 230, 134, 105, 13, 127, 44, 250, 1, 224, 47, 241, 108, 120, 95, 125, 49, 191, 125, 135, 222, 211, 120, 82, 31, 103, 199, 193, 217, 50, 34, 78, 214, 131, 99, 95, 18, 235, 235, 180, 40, 33, 188, 178, 39, 143, 147, 167, 96, 78, 150, 248, 165, 91, 10, 138, 102, 214, 206, 176, 200, 85, 185, 53, 121, 76, 116, 151, 119, 155, 76, 16, 211, 193, 184, 250, 202, 83, 91, 147, 87, 31, 234, 191, 1, 114, 192, 255, 105, 110, 14, 98, 57, 110, 87, 100, 154, 188, 38, 84, 87, 137, 200, 60, 51, 84, 216, 201, 82, 11, 170, 16, 52, 154, 80, 2, 82, 229, 242, 220, 255, 179, 87, 61, 154, 8, 88, 8, 107, 15, 130, 189, 156, 210, 66, 119, 158, 22, 164, 100, 166, 125, 222, 189, 140, 149, 138, 224, 105, 95, 112, 255, 126, 180, 47, 154, 161, 172, 224, 168, 68, 205, 1, 82, 65, 119, 220, 239, 199, 105, 36, 211, 130, 17, 17, 103, 221, 81, 251, 40, 174, 56, 32, 146, 64, 32, 181, 80, 209, 211, 86, 83, 153, 68, 131, 145, 200, 112, 162, 8, 165, 245, 11, 186, 213, 79, 2, 56, 69, 144, 62, 66, 63, 226, 226, 183, 23, 230, 176, 191, 150, 200, 66, 1, 221, 85, 140, 19, 251, 66, 183, 61, 235, 12, 47, 212, 153, 66, 125, 132, 214, 184, 218, 185, 125, 118, 61, 102, 15, 180, 44, 230, 134, 105, 13, 127, 44, 250, 1, 224, 47, 241, 108, 120, 95, 125, 49, 191, 125, 135, 222, 211, 120, 82, 31, 103, 199, 193, 217, 50, 34, 78, 214, 131, 99, 95, 18, 235, 235, 180, 40, 33, 188, 178, 39, 143, 147, 167, 96, 78, 150, 248, 165, 91, 10, 138, 102, 214, 206, 176, 200, 85, 185, 53, 121, 76, 116, 151, 119, 155, 76, 16, 211, 193, 184, 250, 202, 83, 91, 147, 87, 31, 234, 191, 1, 114, 192, 255, 105, 110, 14, 98, 57, 110, 87, 100, 154, 188, 38, 84, 87, 137, 200, 60, 51, 84, 216, 201, 82, 11, 170, 16, 52}
		txs := make([]*ttype.Transaction, 0)

		// fmt.Println(batchsize)
		for i := 0; i < batchsize; i++ {
			tx := &ttype.Transaction{
				Constbytes: messageconst,
			}
			txs = append(txs, tx)
		}

		batchtx := &ttype.BatchTx{
			Txs:       txs,
			Timestamp: time.Now().UnixMilli(),
		}
		// fmt.Println("generated batch size:", len(txs))
		// fmt.Println("generate a batch of txs")
		tp.pending.mu.Lock()
		tp.pending.batchtx = append(tp.pending.batchtx, *batchtx)
		tp.pending.mu.Unlock()

	}

}

func (tp *TxPool) ExtractTxs() *ttype.BatchTx {
	tp.pending.mu.Lock()
	defer tp.pending.mu.Unlock()
	if len(tp.pending.batchtx) == 0 {
		return nil
	}
	batchtx := tp.pending.batchtx[0]
	tp.pending.batchtx = tp.pending.batchtx[1:]
	return &batchtx
}

// GetScale gets the current workload scale
// func (tp *TxPool) GetScale() int {
// 	return tp.pending.len()
// }

// RetrievePending retrieves txs from the pending pool and adds them into the queue pool
// func (tp *TxPool) RetrievePending() {
// 	pendingTxs := tp.pending.empty()
// 	fmt.Printf("tx pool size: %d\n", len(pendingTxs))
// 	for _, t := range pendingTxs {
// 		id := string(t.ID)
// 		tp.queue.Store(id, t)
// 	}
// }

// Pick randomly picks #size txs into a new block
// func (tp *TxPool) Pick(batchSize int) []*ttype.Transaction {

// 	var txs []*ttype.Transaction
// 	var pickSize int
// 	var poolSize int

// 	poolSize = tp.pending.len()

// 	// if the current pool length is smaller than batchsize, then pick all the txs in the pool
// 	if poolSize < batchSize {
// 		pickSize = poolSize
// 	} else {
// 		pickSize = batchSize
// 	}

// 	txs = tp.ExtractTxs(pickSize)
// 	return txs
// }

// func (tp *TxPool) ExtractTxs(size int) []*ttype.Transaction {
// 	tp.pending.mu.Lock()
// 	txs := tp.pending.txs[:size]
// 	tp.pending.mu.Unlock()
// 	return txs
// }

// // Pick randomly picks #size txs into a new block
// func (tp *TxPool) Pick(size int) []*ttype.Transaction {
// 	var selectedTxs = make(map[int]struct{})
// 	var ids []string
// 	var txs []*ttype.Transaction

// 	tp.queue.Range(func(key, value any) bool {
// 		id := key.(string)
// 		ids = append(ids, id)
// 		return true
// 	})

// 	for i := 0; i < size; i++ {
// 		for {
// 			rand.Seed(time.Now().UnixNano())
// 			random := rand.Intn(len(ids))
// 			if _, ok := selectedTxs[random]; !ok {
// 				selectedTxs[random] = struct{}{}
// 				break
// 			}
// 		}
// 	}

// 	for key := range selectedTxs {
// 		id := ids[key]
// 		v, _ := tp.queue.Load(id)
// 		t, ok := v.(*ttype.Transaction)
// 		if ok {
// 			txs = append(txs, t)
// 		}
// 	}

// 	return txs
// }

// DeleteTxs deletes txs in other concurrent blocks
// func (tp *TxPool) DeleteTxs(txs []*ttype.Transaction) []*ttype.Transaction {
// 	var deleted []*ttype.Transaction

// 	for _, del := range txs {
// 		id := string(del.ID)
// 		if v, ok := tp.queue.Load(id); ok {
// 			t := v.(*ttype.Transaction)
// 			deleted = append(deleted, t)
// 			tp.queue.Delete(id)
// 		}
// 	}

// 	return deleted
// }

type pendingPool struct {
	mu sync.RWMutex
	//txs map[string]*types.Transaction
	batchtx []ttype.BatchTx
}

// func (pp *pendingPool) Append(tx *ttype.Transaction) {
// 	//id := string(tx.ID)
// 	//pp.txs[id] = tx
// 	pp.mu.Lock()
// 	defer pp.mu.Unlock()
// 	pp.txs = append(pp.txs, tx)
// }

// func (pp *pendingPool) BatchAppend(txs []*ttype.Transaction) {
// 	for _, t := range txs {
// 		pp.Append(t)
// 	}
// }

// func (pp *pendingPool) pop() *ttype.Transaction {
// 	pp.mu.Lock()
// 	defer pp.mu.Unlock()
// 	last := pp.txs[len(pp.txs)-1]
// 	pp.txs = pp.txs[:len(pp.txs)-1]
// 	return last
// }

// func (pp *pendingPool) delete(tx *ttype.Transaction) {
// 	pp.mu.Lock()
// 	defer pp.mu.Unlock()
// 	for i := range pp.txs {
// 		if pp.txs[i] == tx {
// 			pp.txs = append(pp.txs[:i], pp.txs[i+1:]...)
// 			break
// 		}
// 	}
// 	//id := string(tx.ID)
// 	//delete(pp.txs, id)
// }

// func (pp *pendingPool) len() int {
// 	pp.mu.RLock()
// 	defer pp.mu.RUnlock()
// 	return len(pp.txs)
// }

// func (pp *pendingPool) swap(i, j int) {
// 	pp.mu.Lock()
// 	defer pp.mu.Unlock()
// 	pp.txs[i], pp.txs[j] = pp.txs[j], pp.txs[i]
// }

// func (pp *pendingPool) isFull() bool {
// 	pp.mu.RLock()
// 	defer pp.mu.RUnlock()
// 	return len(pp.txs) == MaximumPoolSize
// }

// func (pp *pendingPool) isEmpty() bool {
// 	return len(pp.txs) == 0
// }

// func (pp *pendingPool) empty() []*ttype.Transaction {
// 	pp.mu.Lock()
// 	defer pp.mu.Unlock()

// 	var txs []*ttype.Transaction
// 	if !pp.isEmpty() {
// 		//for _, t := range pp.txs {
// 		//	txs = append(txs, t)
// 		//}
// 		txs = pp.txs
// 		pp.txs = make([]*ttype.Transaction, 0, MaximumPoolSize)
// 	}

// 	return txs
// }

// // func (pp *pendingPool) existInPending(tx *ttype.Transaction) bool {
// // 	id := string(tx.ID)
// // 	for _, t := range pp.txs {
// // 		if strings.Compare(id, string(t.ID)) == 0 {
// // 			return true
// // 		}
// // 	}
// // 	return false
// // }

func newPending() *pendingPool {
	return &pendingPool{
		batchtx: make([]ttype.BatchTx, 0, MaximumPoolSize),

		//txs: make(map[string]*types.Transaction),
	}
}
