package state

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// peerUpdateInterval represents the interval of finding new peer nodes
// and updating the blockchain on disk with missing blocks.
const peerUpdateInterval = time.Minute

// worker manages the POW workflows for the blockchain.
type worker struct {
	state        *State
	wg           sync.WaitGroup
	ticker       time.Ticker
	shut         chan struct{}
	peerUpdates  chan bool
	startMining  chan bool
	cancelMining chan chan struct{}
	evHandler    EventHandler
	baseURL      string
}

// runWorker creates a powWorker for starting the POW workflows.
func runWorker(state *State, evHandler EventHandler) {

	// Construct and register this worker to the state. During initialization
	// this worker needs access to the state.
	state.worker = &worker{
		state:        state,
		ticker:       *time.NewTicker(peerUpdateInterval),
		shut:         make(chan struct{}),
		peerUpdates:  make(chan bool, 1),
		startMining:  make(chan bool, 1),
		cancelMining: make(chan chan struct{}, 1),
		evHandler:    evHandler,
		baseURL:      "http://%s/v1/node",
	}

	// Update this node before starting any support G's.
	state.worker.sync()

	// Load the set of operations we need to run.
	operations := []func(){
		state.worker.miningOperations,
	}

	// Set waitgroup to match the number of G's we need for the set
	// of operations we have.
	g := len(operations)
	state.worker.wg.Add(g)

	// We don't want to return until we know all the G's are up and running.
	hasStarted := make(chan bool)

	// Start all the operational G's.
	for _, op := range operations {
		go func(op func()) {
			defer state.worker.wg.Done()
			hasStarted <- true
			op()
		}(op)
	}

	// Wait for the G's to report they are running.
	for i := 0; i < g; i++ {
		<-hasStarted
	}
}

// shutdown terminates the goroutine performing work.
func (w *worker) shutdown() {
	w.evHandler("worker: shutdown: started")
	defer w.evHandler("worker: shutdown: completed")

	w.evHandler("worker: shutdown: signal cancel mining")
	done := w.signalCancelMining()
	done()

	w.evHandler("worker: shutdown: terminate goroutines")
	close(w.shut)
	w.wg.Wait()
}

// =============================================================================

// sync updates the peer list, mempool and blocks.
func (w *worker) sync() {
	w.evHandler("worker: sync: started")
	defer w.evHandler("worker: sync: completed")

	for _, peer := range w.state.RetrieveKnownPeers() {

		// Retrieve the status of this peer.
		peerStatus, err := w.queryPeerStatus(peer)
		if err != nil {
			w.evHandler("worker: sync: queryPeerStatus: %s: ERROR: %s", peer.Host, err)
		}

		// Add new peers to this nodes list.
		if err := w.addNewPeers(peerStatus.KnownPeers); err != nil {
			w.evHandler("worker: sync: addNewPeers: %s: ERROR: %s", peer.Host, err)
		}

		// Update the mempool.
		pool, err := w.queryPeerMempool(peer)
		if err != nil {
			w.evHandler("worker: sync: queryPeerMempool: %s: ERROR: %s", peer.Host, err)
		}
		for _, tx := range pool {
			w.evHandler("worker: sync: queryPeerMempool: %s: Add Tx: %s", peer.Host, tx.SignatureString()[:16])
			w.state.mempool.Upsert(tx)
		}

		// If this peer has blocks we don't have, we need to add them.
		if peerStatus.LatestBlockNumber > w.state.RetrieveLatestBlock().Header.Number {
			w.evHandler("worker: sync: writePeerBlocks: %s: latestBlockNumber[%d]", peer.Host, peerStatus.LatestBlockNumber)
			if err := w.writePeerBlocks(peer); err != nil {
				w.evHandler("worker: sync: writePeerBlocks: %s: ERROR %s", peer.Host, err)
			}
		}
	}
}

// =============================================================================

// queryPeerStatus looks for new nodes on the blockchain by asking
// known nodes for their peer list. New nodes are added to the list.
func (w *worker) queryPeerStatus(pr peer.Peer) (peer.PeerStatus, error) {
	w.evHandler("worker: runPeerUpdatesOperation: queryPeerStatus: started: %s", pr)
	defer w.evHandler("worker: runPeerUpdatesOperation: queryPeerStatus: completed: %s", pr)

	url := fmt.Sprintf("%s/status", fmt.Sprintf(w.baseURL, pr.Host))

	var ps peer.PeerStatus
	if err := send(http.MethodGet, url, nil, &ps); err != nil {
		return peer.PeerStatus{}, err
	}

	w.evHandler("worker: runPeerUpdatesOperation: queryPeerStatus: peer-node[%s]: latest-blknum[%d]: peer-list[%s]", pr, ps.LatestBlockNumber, ps.KnownPeers)

	return ps, nil
}

// queryPeerMempool asks the peer for their current copy of their mempool
func (w *worker) queryPeerMempool(pr peer.Peer) ([]storage.BlockTx, error) {
	w.evHandler("worker: runPeerUpdatesOperation: queryPeerMempool: started: %s", pr)
	defer w.evHandler("worker: runPeerUpdatesOperation: queryPeerMempool: completed: %s", pr)

	url := fmt.Sprintf("%s/tx/list", fmt.Sprintf(w.baseURL, pr.Host))

	var mempool []storage.BlockTx
	if err := send(http.MethodGet, url, nil, &mempool); err != nil {
		return nil, err
	}

	w.evHandler("worker: runPeerUpdatesOperation: queryPeerMempool: len[%d]", len(mempool))

	return mempool, nil
}

// addNewPeers takes the list of known peers and makes sure they are included
// in the nodes list of know peers.
func (w *worker) addNewPeers(knownPeers []peer.Peer) error {
	w.evHandler("worker: runPeerUpdatesOperation: addNewPeers: started")
	defer w.evHandler("worker: runPeerUpdatesOperation: addNewPeers: completed")

	for _, peer := range knownPeers {
		if err := w.state.addPeerNode(peer); err != nil {

			// It already exists, nothing to report.
			return nil
		}
		w.evHandler("worker: runPeerUpdatesOperation: addNewPeers: add peer nodes: adding peer-node %s", peer)
	}

	return nil
}

// writePeerBlocks queries the specified node asking for blocks this
// node does not have, then writes them to disk.
func (w *worker) writePeerBlocks(pr peer.Peer) error {
	w.evHandler("worker: runPeerUpdatesOperation: writePeerBlocks: started: %s", pr)
	defer w.evHandler("worker: runPeerUpdatesOperation: writePeerBlocks: completed: %s", pr)

	from := w.state.RetrieveLatestBlock().Header.Number + 1
	url := fmt.Sprintf("%s/block/list/%d/latest", fmt.Sprintf(w.baseURL, pr.Host), from)

	var blocks []storage.Block
	if err := send(http.MethodGet, url, nil, &blocks); err != nil {
		return err
	}

	w.evHandler("worker: runPeerUpdatesOperation: writePeerBlocks: found blocks[%d]", len(blocks))

	for _, block := range blocks {
		w.evHandler("worker: runPeerUpdatesOperation: writePeerBlocks: prevBlk[%s]: newBlk[%s]: numTrans[%d]", block.Header.ParentHash, block.Hash(), len(block.Transactions))

		if err := w.state.WritePeerBlock(block); err != nil {
			return err
		}
	}

	return nil
}

// =============================================================================

// miningOperations handles mining.
func (w *worker) miningOperations() {
	w.evHandler("worker: miningOperations: G started")
	defer w.evHandler("worker: miningOperations: G completed")

	for {
		select {
		case <-w.startMining:
			if !w.isShutdown() {
				w.runMiningOperation()
			}
		case <-w.shut:
			w.evHandler("worker: miningOperations: received shut signal")
			return
		}
	}
}

// isShutdown is used to test if a shutdown has been signaled.
func (w *worker) isShutdown() bool {
	select {
	case <-w.shut:
		return true
	default:
		return false
	}
}

// =============================================================================

// signalStartMining starts a mining operation. If there is already a signal
// pending in the channel, just return since a mining operation will start.
func (w *worker) signalStartMining() {
	select {
	case w.startMining <- true:
	default:
	}
	w.evHandler("worker: signalStartMining: mining signaled")
}

// signalCancelMining signals the G executing the runMiningOperation function
// to stop immediately. That G will not return from the function until done
// is called. This allows the caller to complete any state changes before a new
// mining operation takes place.
func (w *worker) signalCancelMining() (done func()) {
	wait := make(chan struct{})

	select {
	case w.cancelMining <- wait:
	default:
	}
	w.evHandler("worker: signalCancelMining: cancel mining signaled")

	return func() { close(wait) }
}

// =============================================================================

// runMiningOperation takes all the transactions from the mempool and writes a
// new block to the database.
func (w *worker) runMiningOperation() {
	w.evHandler("worker: runMiningOperation: MINING: started")
	defer w.evHandler("worker: runMiningOperation: MINING: completed")

	// Make sure there are at least transPerBlock in the mempool.
	length := w.state.QueryMempoolLength()
	if length < w.state.genesis.TransPerBlock {
		w.evHandler("worker: runMiningOperation: MINING: not enough transactions to mine: Txs[%d]", length)
		return
	}

	// After running a mining operation, check if a new operation should
	// be signaled again.
	defer func() {
		length := w.state.QueryMempoolLength()
		if length >= w.state.genesis.TransPerBlock {
			w.evHandler("worker: runMiningOperation: MINING: signal new mining operation: Txs[%d]", length)
			w.signalStartMining()
		}
	}()

	// If mining is signalled to be cancelled by the WriteNextBlock function,
	// this G can't terminate until it is told it can.
	var wait chan struct{}
	defer func() {
		if wait != nil {
			w.evHandler("worker: runMiningOperation: MINING: termination signal: waiting")
			<-wait
			w.evHandler("worker: runMiningOperation: MINING: termination signal: received")
		}
	}()

	// Drain the cancel mining channel before starting.
	select {
	case <-w.cancelMining:
		w.evHandler("worker: runMiningOperation: MINING: drained cancel channel")
	default:
	}

	// Create a context so mining can be cancelled.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Can't return from this function until these G's are complete.
	var wg sync.WaitGroup
	wg.Add(2)

	// This G exists to cancel the mining operation.
	go func() {
		defer func() {
			cancel()
			wg.Done()
		}()

		select {
		case wait = <-w.cancelMining:
			w.evHandler("worker: runMiningOperation: MINING: cancel mining requested")
		case <-ctx.Done():
		}
	}()

	// This G is performing the mining.
	go func() {
		defer func() {
			cancel()
			wg.Done()
		}()

		_, duration, err := w.state.MineNewBlock(ctx)
		w.evHandler("worker: runMiningOperation: MINING: mining duration[%v]", duration)

		if err != nil {
			switch {
			case errors.Is(err, ErrNotEnoughTransactions):
				w.evHandler("worker: runMiningOperation: MINING: WARNING: not enough transactions in mempool")
			case ctx.Err() != nil:
				w.evHandler("worker: runMiningOperation: MINING: CANCELLED: by request")
			default:
				w.evHandler("worker: runMiningOperation: MINING: ERROR: %s", err)
			}
			return
		}

		// WOW, we mined a block. Send the new block to the network.
		// Log the error, but that's it.
	}()

	// Wait for both G's to terminate.
	wg.Wait()
}

// =============================================================================

// send is a helper function to send an HTTP request to a node.
func send(method string, url string, dataSend interface{}, dataRecv interface{}) error {
	var req *http.Request

	switch {
	case dataSend != nil:
		data, err := json.Marshal(dataSend)
		if err != nil {
			return err
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(data))
		if err != nil {
			return err
		}

	default:
		var err error
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return err
		}
	}

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(msg))
	}

	if dataRecv != nil {
		if err := json.NewDecoder(resp.Body).Decode(dataRecv); err != nil {
			return err
		}
	}

	return nil
}
