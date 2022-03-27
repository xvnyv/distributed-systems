package lib

import (
	"encoding/json"
	"fmt"
	"log"

	badger "github.com/dgraph-io/badger/v3"
)

func (n *Node) BadgerWrite(o []ClientCart) error {
	opts := badger.DefaultOptions(fmt.Sprintf("tmp/%v/badger", n.Id))
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		log.Printf("Badger Error: %v\n", err)
		return err
	}
	defer db.Close()

	for _, v := range o {
		toWrite := ClientCart{}
		// INIT for reading ----
		res := ClientCart{}
		err = db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(v.UserID))
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
			//reading complete here ----
		})

		if err != nil {
			if err.Error() == "Key not found" {
				toWrite = v
				fmt.Printf("Key not found, writing %v\n", v)
				//do nothing
			} else {
				return err
			}
		} else {
			//check whether current vector clock smaller than received
			fmt.Printf("found previous verison: %v\n", res.VectorClock)
			fmt.Printf("comparing with new version: %v\n", v.VectorClock)

			if VectorClockSmaller(res.VectorClock, v.VectorClock) {
				fmt.Println("current value in db vector clock smaller than new write val: Overwrite")
				toWrite = v
			} else {
				fmt.Println("current value in db vector clock vs new write val ambiguos: Merge")
				toWrite = MergeClientCarts(res, v)
			}
		}

		err = db.Update(func(txn *badger.Txn) error {
			//need convert DataObject to byte array
			//forloop
			if v.UserID == "" {
				log.Println("No UserId. Object is:", toWrite)
			}
			dataObjectBytes, _ := json.Marshal(toWrite)
			err := txn.Set([]byte(toWrite.UserID), dataObjectBytes)
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
func (n *Node) BadgerRead(key string) (ClientCart, error) {
	opts := badger.DefaultOptions(fmt.Sprintf("tmp/%v/badger", n.Id))
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		log.Printf("Badger Error: %v\n", err)
		return ClientCart{}, err
	}
	defer db.Close()
	// Your code here…

	res := ClientCart{}
	err = db.View(func(txn *badger.Txn) error {
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
	opts := badger.DefaultOptions(fmt.Sprintf("tmp/%v/badger", n.Id))
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		log.Printf("Badger Error: %v\n", err)
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
