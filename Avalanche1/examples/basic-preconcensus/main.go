package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"

	avalanche "github.com/tyler-smith/go-avalanche"
)

const (
	nodeCount = 1000
	zc_node   = 1000 - ey_node
	ey_node   = 300
	txCount   = 1000
	r_p       = 0.1
)

var (
	networkNodes   []*node
	loggingEnabled = true
)

func main() {

	count := 0  //通信次数计数
	count1 := 0 //通信次数计数
	logging := flag.Bool("logging", false, "Enable logging")
	flag.Parse()

	if logging != nil {
		loggingEnabled = *logging
	}

	// Create nodes
	networkNodes = make([]*node, nodeCount)
	for i := 0; i < nodeCount; i++ {
		networkNodes[i] = newNode(avalanche.NodeID(i), avalanche.NewConnman())
	}
	t1 := time.Now()
	// Create wg with a slot for each node
	wg := &sync.WaitGroup{}
	wg.Add(nodeCount)

	// Start node processing with wg to signal completion
	for i := 0; i < nodeCount; i++ {
		go networkNodes[i].run(wg)
		count++

	}

	t0 := time.Now()

	// Send txs to each node
	/*
		for _, t := range rand.Perm(txCount) {
			for i := 0; i < nodeCount; i++ {
				networkNodes[i].incoming <- &tx{hash: int64(t), isAccepted: true}
			}
			count = count + (t - 1)
			count1 = count1 + (t - 1)
		}
	*/
	for _, t := range rand.Perm(txCount) {
		for i := 0; i < zc_node; i++ {
			networkNodes[i].incoming <- &tx{hash: int64(t), isAccepted: true}
		}
		count = count + (t - 1)
		count1 = count1 + (t - 1)
	}
	for _, t := range rand.Perm(txCount) {
		for i := 0; i < ey_node; i++ {
			networkNodes[i].incoming <- &tx{hash: int64(t), isAccepted: false}
		}
		count = count + (t - 1)
		count1 = count1 + (t - 1)
	}
	time.Sleep(time.Duration(1000*r_p*txCount/count1) * time.Millisecond) //r+p p:事件传播时间 r：节点周期性地生成事件
	// Stop all nodes
	for i := 0; i < nodeCount; i++ {
		close(networkNodes[i].incoming)
		count++
	}

	// Wait for all nodes to finish
	wg.Wait()

	t2 := time.Now()
	fmt.Println(fmt.Sprintf("Finished in %fs", time.Now().Sub(t0).Seconds()))
	fmt.Println("Nodes fully finalized: ", nodesFullyFinalized)

	//TPS吞吐量  定义：上链的总交易数/其区块生成时间生  成时间包括交易+共识机制执行+区块
	tps := txCount * r_p / (0.25 * (time.Now().Sub(t0).Seconds() - time.Now().Sub(t2).Seconds()) / (1000 + txCount - zc_node)) * (1 + ey_node*0.01)
	fmt.Println("TPS: ")
	fmt.Println(tps)

	//delay 时间延迟 定义：交易发出到确认的时间 平均值
	delay := (time.Now().Sub(t1).Seconds() - time.Now().Sub(t2).Seconds()) * float64(zc_node) / txCount / (1 + ey_node*0.006)
	delay = (delay + float64(zc_node)*0.006 + r_p*0.5) / r_p
	fmt.Println("Delay: ")
	fmt.Println(delay)

	//平均通信开销 定义为节点间的通信次数
	exchange := float64(count) / (zc_node / 2)
	fmt.Println("Exchange: ")
	fmt.Println(exchange)
}

func log(str string, args ...interface{}) {
	if loggingEnabled {
		fmt.Println(fmt.Sprintf(str, args...))
	}
}

type node struct {
	id         avalanche.NodeID
	snowball   *avalanche.Processor
	snowballMu *sync.RWMutex
	incoming   chan (*tx)
}

func newNode(id avalanche.NodeID, connman *avalanche.Connman) *node {
	return &node{
		id:         id,
		snowball:   avalanche.NewProcessor(connman),
		snowballMu: &sync.RWMutex{},
		incoming:   make(chan (*tx), 10),
	}
}

var nodesFullyFinalized = 0

func (n node) run(wg *sync.WaitGroup) {
	defer wg.Done()

	finalizedCount := 0

	// HACK: figure out how to best signal event loop so it quits if and only if
	// we're done adding and all txs are finalized
	doneAdding := make(chan (struct{}))
	go func() {
		for t := range n.incoming {
			n.snowballMu.Lock()
			n.snowball.AddTargetToReconcile(t)
			n.snowballMu.Unlock()
		}
		close(doneAdding)
	}()
	<-doneAdding

	queries := 0
	for i := 0; i < 1e8; i++ {
		nodeID := i % len(networkNodes)

		// Don't query ourself
		if nodeID == int(n.id) {
			continue
		}

		queries++
		updates := []avalanche.StatusUpdate{}

		// Query node
		n.snowballMu.Lock()
		invs := n.snowball.GetInvsForNextPoll()
		n.snowballMu.Unlock()

		// All done
		// if len(invs) == 0 {
		// 	fmt.Println("Out of invs:", n.id)
		// 	return
		// }

		resp := networkNodes[nodeID].query(invs)

		// Register query response
		n.snowballMu.Lock()
		n.snowball.RegisterVotes(n.id, resp, &updates)
		n.snowballMu.Unlock()

		if len(updates) == 0 {
			continue
		}

		for _, update := range updates {
			if update.Status == avalanche.StatusFinalized {
				finalizedCount++
				fmt.Println("Finalized tx %d on node %d after %d queries", update.Hash, n.id, queries)
			} else if update.Status == avalanche.StatusAccepted {
				fmt.Println("Accepted tx %d on node %d after %d queries", update.Hash, n.id, queries)
			} else if update.Status == avalanche.StatusRejected {
				fmt.Println("Rejected tx %d on node %d after %d queries", update.Hash, n.id, queries)
			} else if update.Status == avalanche.StatusInvalid {
				fmt.Println("Invalidated tx %d on node %d after %d queries", update.Hash, n.id, queries)
			} else {
				fmt.Println(update.Status == avalanche.StatusAccepted)
				panic(update)
			}
		}

		if finalizedCount >= txCount {
			nodesFullyFinalized++
			return
		}
	}

	log("Limit exceeded")
}

func (n node) query(invs []avalanche.Inv) avalanche.Response {
	n.snowballMu.Lock()
	defer n.snowballMu.Unlock()

	votes := make([]avalanche.Vote, len(invs))
	for i := 0; i < len(invs); i++ {

		t := &tx{hash: int64(invs[i].TargetHash), isAccepted: true}

		n.snowball.AddTargetToReconcile(t)

		var vote uint32 = 0
		if !n.snowball.IsAccepted(t) {
			vote = 1
		}

		// Randomly flip votes to prolong convergence
		// if rand.Float64()*100 < 35 {
		// 	vote = vote ^ 1
		// }

		votes[i] = avalanche.NewVote(vote, invs[i].TargetHash)
	}

	return avalanche.NewResponse(0, 0, votes)
}

// tx：表示一个交易，具有哈希值和接受状态
type tx struct {
	hash       int64
	isAccepted bool
}

func (t *tx) Hash() avalanche.Hash { return avalanche.Hash(t.hash) }

func (t *tx) IsAccepted() bool { return t.isAccepted }

func (*tx) IsValid() bool { return true }

func (*tx) Type() string { return "tx" }

func (*tx) Score() int64 { return 1 }
