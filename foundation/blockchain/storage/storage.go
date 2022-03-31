package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Storage manages reading and writing of blocks to storage.
type Storage struct {
	dbPath string
	dbFile *os.File
	mu     sync.Mutex
}

// New provides access to blockchain storage.
func New(dbPath string) (*Storage, error) {

	// Open the blockchain database file with append.
	dbFile, err := os.OpenFile(dbPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	strg := Storage{
		dbPath: dbPath,
		dbFile: dbFile,
	}

	return &strg, nil
}

// Close cleanly releases the storage area.
func (str *Storage) Close() {
	str.mu.Lock()
	defer str.mu.Unlock()

	str.dbFile.Close()
}

// Write adds a new block to the chain.
func (str *Storage) Write(block BlockFS) error {
	str.mu.Lock()
	defer str.mu.Unlock()

	// Marshal the block for writing to disk.
	blockFSJson, err := json.Marshal(block)
	if err != nil {
		return err
	}

	// Write the new block to the chain on disk.
	if _, err := str.dbFile.Write(append(blockFSJson, '\n')); err != nil {
		return err
	}

	return nil
}

// =============================================================================

// ReadAllBlocks loads all existing blocks from storage into memory. In a real
// world situation this would require a lot of memory.
func (str *Storage) ReadAllBlocks() ([]Block, error) {
	dbFile, err := os.Open(str.dbPath)
	if err != nil {
		return nil, err
	}
	defer dbFile.Close()

	var blockNum int
	var blocks []Block
	scanner := bufio.NewScanner(dbFile)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var blockFS BlockFS
		if err := json.Unmarshal(scanner.Bytes(), &blockFS); err != nil {
			return nil, err
		}

		if blockFS.Block.Hash() != blockFS.Hash {
			return nil, fmt.Errorf("block %d has been changed", blockNum)
		}

		blocks = append(blocks, blockFS.Block)
		blockNum++
	}

	return blocks, nil
}
