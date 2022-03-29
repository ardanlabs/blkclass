package storage

import (
	"encoding/json"
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
