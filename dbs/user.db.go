package dbs

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/fxamacker/cbor/v2"
)

// DB Methods
func NewUserTestDB(db *badger.DB) *UserTestDB {
	return &UserTestDB{
		db:     db,
		prefix: []byte("usere-"),
	}
}

func (h *UserTestDB) Save(username string, usere UserTestDBEntry) error {
	username = strings.ToLower(username)

	usereBytes, err := cbor.Marshal(usere)
	if err != nil {
		return errors.New("Failed to marshal User entry. The error is: " + err.Error())
	}

	userEStorageId := append(h.prefix, []byte(username)...)

	dbtxn := h.db.NewTransaction(true)
	defer dbtxn.Discard()

	entry := badger.NewEntry(userEStorageId, usereBytes)
	err = dbtxn.SetEntry(entry)
	if err != nil {
		return errors.New("Failed creating User db entry instance. The error is: " + err.Error())
	}

	err = dbtxn.Commit()
	if err != nil {
		return errors.New("Failed saving User entry. The error is: " + err.Error())
	}

	return nil
}

func (h *UserTestDB) Get(username string) (*UserTestDBEntry, error) {
	username = strings.ToLower(username)

	userEStorageId := append(h.prefix, []byte(username)...)

	dbtxn := h.db.NewTransaction(true)
	defer dbtxn.Discard()

	item, err := dbtxn.Get(userEStorageId)
	if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
		return nil, fmt.Errorf("The entry with id %s does not exist", username)
	} else if err != nil {
		return nil, errors.New("Failed locating entry. The error is: " + err.Error())
	}

	itemBytes, err := item.ValueCopy(nil)
	if err != nil {
		return nil, errors.New("Failed reading entry value. The error is: " + err.Error())
	}

	var usertEntryInst UserTestDBEntry
	err = cbor.Unmarshal(itemBytes, &usertEntryInst)
	if err != nil {
		return nil, errors.New("Failed cbor decoding entry value. The error is: " + err.Error())
	}

	return &usertEntryInst, nil
}

func (h *UserTestDB) ResetUsers() error {
	dbtxn := h.db.NewTransaction(true)
	defer dbtxn.Discard()

	iterTxn := dbtxn.NewIterator(badger.IteratorOptions{
		Prefix: h.prefix,
	})
	defer iterTxn.Close()
	for iterTxn.Rewind(); iterTxn.Valid(); iterTxn.Next() {
		item := iterTxn.Item()
		k := item.Key()

		log.Println("Deleting... " + hex.EncodeToString(k))

		err := dbtxn.Delete(k)
		if err != nil {
			log.Println("Error creater delete req... " + hex.EncodeToString(k))
		}

		err = dbtxn.Commit()
		if err != nil {
			log.Println("Failed to commit delete req... " + hex.EncodeToString(k))
		}
	}

	return nil
}
