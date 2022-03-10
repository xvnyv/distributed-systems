package lib

import (
	"encoding/json"
	"fmt"
	"log"

	badger "github.com/dgraph-io/badger/v3"
)

func (n *Node) BadgerWrite(o []DataObject) error {
	opts := badger.DefaultOptions(fmt.Sprintf("tmp/%v/badger", n.Id))
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer db.Close()
	// Your code here…
	for _, v := range o {
		err = db.Update(func(txn *badger.Txn) error {
			//need convert DataObject to byte array
			//forloop
			if v.Key == "" {
				fmt.Println(v)
			}
			dataObjectBytes, _ := json.Marshal(v)
			err := txn.Set([]byte(v.Key), dataObjectBytes)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

/**
Returns empty DataObject if there is an error reading from the database with the provided key.
*/
func (n *Node) BadgerRead(key string) (DataObject, error) {
	opts := badger.DefaultOptions(fmt.Sprintf("tmp/%v/badger", n.Id))
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
		return DataObject{}, err
	}
	defer db.Close()
	// Your code here…

	res := DataObject{}
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		// Alternatively, you could also use item.ValueCopy().
		// valCopy, err := item.ValueCopy(nil)
		// handle(err)
		//
		var valCopy []byte

		err = item.Value(func(val []byte) error {
			// This func with val would only be called if item.Value encounters no error.

			// Copying or parsing val is valid.
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
	opts := badger.DefaultOptions(fmt.Sprintf("tmp/%v/badger", n.Id))
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer db.Close()
	// Your code here…
	for _, v := range keys {
		err = db.Update(func(txn *badger.Txn) error {
			err := txn.Delete([]byte(v))
			return err
		})
		if err != nil {
			return err
		}
	}

	return err
}

func (n *Node) BadgerGetKeys() ([]string, error) {
	opts := badger.DefaultOptions(fmt.Sprintf("tmp/%v/badger", n.Id))
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return []string{}, err
	}
	defer db.Close()
	result := make([]string, 0)

	err = db.View(func(txn *badger.Txn) error {
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
