package lib

import (
	"encoding/json"
	"fmt"
	"log"

	badger "github.com/dgraph-io/badger/v3"
)

var db *badger.DB

func (n *Node) OpenBadger() {
	opts := badger.DefaultOptions(fmt.Sprintf("tmp/%v/badger", n.Id))
	opts.Logger = nil
	opts.SyncWrites = true

	openedDb, err := badger.Open(opts)
	for i := 0; i < DB_OPEN_RETIRES && err != nil; i++ {
		openedDb, err = badger.Open(opts)
	}
	if err != nil {
		log.Fatalf("Error opening DB: %s\n", err)
	}
	db = openedDb
}

func (n *Node) CloseBadger() {
	if db != nil {
		log.Println("Closing DB")
		defer db.Close()
	}
}

func (n *Node) BadgerWrite(c ClientCart) (BadgerObject, error) {
	toWrite := BadgerObject{}
	userid := c.UserID
	// if conflict, no need to read
	conflict := false
	lastWritten, err := n.BadgerRead(userid)
	if err != nil {
		if err.Error() == "Key not found" {
			toWrite = BadgerObject{UserID: userid, Versions: []ClientCart{c}}
			fmt.Printf("Key not found, writing %v\n", toWrite)
			//do nothing
		} else {
			return BadgerObject{}, err // Error, return empty client cart
		}
	} else {
		//iterate through the versions, check whether can overwrite
		newVersions := []ClientCart{}
		for i := 0; i < len(lastWritten.Versions); i++ {
			if VectorClockSmaller(lastWritten.Versions[i].VectorClock, c.VectorClock) {
				// if current version is smaller than incoming version
				// don't add to new array of versions
				continue
			}
			//add to new array of versions
			newVersions = append(newVersions, lastWritten.Versions[i])
		}
		newVersions = append(newVersions, c)

		if len(newVersions) > 1 {
			conflict = true
		}

		lastWritten.Versions = newVersions
		toWrite = lastWritten
	}

	err = db.Update(func(txn *badger.Txn) error {
		//need convert DataObject to byte array
		dataObjectBytes, _ := json.Marshal(toWrite)
		err := txn.Set([]byte(toWrite.UserID), dataObjectBytes)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return BadgerObject{}, err
	}

	// NOTE: returning badger object with one version DESPITE badger object possibly
	// having MULTIPLE versions
	return BadgerObject{UserID: userid, Versions: []ClientCart{c}, Conflict: conflict}, nil
}

/**
Returns empty DataObject if there is an error reading from the database with the provided key.
*/
func (n *Node) BadgerRead(key string) (BadgerObject, error) {
	res := BadgerObject{}
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		var valCopy []byte

		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)

			return nil
		})
		if err != nil {
			return err
		}
		//convert valCopy to DataObject
		err = json.Unmarshal(valCopy, &res)
		return err
	})

	return res, err
}

func (n *Node) BadgerDelete(keys []string) error {
	var err error
	err = db.Update(func(txn *badger.Txn) error {
		for _, v := range keys {
			err := txn.Delete([]byte(v))
			if err != nil {
				return err
			}
		}
		return err
	})
	return err
}

func (n *Node) BadgerGetKeys() ([]string, error) {
	result := make([]string, 0)

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			result = append(result, string(k))
		}
		return nil
	})
	return result, err
}

func (n *Node) BadgerMigrateWrite(data []BadgerObject) error {
	err := db.Update(func(txn *badger.Txn) error {
		for _, item := range data {
			//need convert DataObject to byte array
			dataObjectBytes, _ := json.Marshal(item)
			err := txn.Set([]byte(item.UserID), dataObjectBytes)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
