package repository

import "sync"

var SharedDataSource []string
var SharedDataSourceMutex sync.Mutex
