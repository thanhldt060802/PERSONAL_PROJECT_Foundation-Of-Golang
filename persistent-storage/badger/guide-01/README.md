### Document for badger library usage

- func Open(opt Options) (*DB, error)
	This function will initialize a *DB with options.

- func (db *badger.DB) RunValueLogGC(discardRatio float64) error
	This function will start a garbage collector to cleanup file storage if the proportion of data that is no longer useful is greater than discardRatio.

- func (db *badger.DB) Update(fn func(txn *badger.Txn) error) error
	This function will start a transaction for a lot of action and commit changes.

- func (txn *badger.Txn) Set(key []byte, val []byte) error
	This function will set an element with key: value format.

- func (txn *badger.Txn) Delete(key []byte) error
	This function will delete an element by key.

- func (txn *badger.Txn) NewIterator(opt badger.IteratorOptions) *badger.Iterator
	This function will create a iterator for retrieving data in (queue) disk.