package state

import (
	"context"
	"errors"
	"sync"
)

// worker manages the POW workflows for the blockchain.
type worker struct {
	state        *State
	wg           sync.WaitGroup
	shut         chan struct{}
	startMining  chan bool
	cancelMining chan chan struct{}
	evHandler    EventHandler
}

// runWorker creates a powWorker for starting the POW workflows.
func runWorker(state *State, evHandler EventHandler) {

	// Construct and register this worker to the state. During initialization
	// this worker needs access to the state.
	state.worker = &worker{
		state:        state,
		shut:         make(chan struct{}),
		startMining:  make(chan bool, 1),
		cancelMining: make(chan chan struct{}, 1),
		evHandler:    evHandler,
	}

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
