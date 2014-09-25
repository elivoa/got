/*
Package sessions implmented session in got.
Session Objects are stored in memory, all sessions are abandoned when the server restart.
Will generate an session id for each session. Save to client in client which key is 'JSESSIONID'

there is something copied from gorilla/context.
*/

package sessions

import (
	"sync"
	"time"
)

var (
	mutex sync.RWMutex
	// session key is session id.
	data  = make(map[string]map[interface{}]interface{})
	datat = make(map[string]int64)
)

// Set stores a value for a given key in a given session.
func Set(sessionId string, key, val interface{}) {
	mutex.Lock()
	if data[sessionId] == nil {
		data[sessionId] = make(map[interface{}]interface{})
		datat[sessionId] = time.Now().Unix()
	}
	data[sessionId][key] = val
	mutex.Unlock()
}

// Get returns a value stored for a given key in a given request.
func Get(sessionId string, key interface{}) interface{} {
	mutex.RLock()
	if ctx := data[sessionId]; ctx != nil {
		value := ctx[key]
		mutex.RUnlock()
		return value
	}
	mutex.RUnlock()
	return nil
}

func GetString(sessionId string, key interface{}) string {
	if ifValue := Get(sessionId, key); ifValue != nil {
		return ifValue.(string)
	} else {
		return ""
	}
}

func GetInt(sessionId string, key interface{}) int {
	if ifValue := Get(sessionId, key); ifValue != nil {
		return ifValue.(int)
	} else {
		return 0
	}
}

// GetOk returns stored value and presence state like multi-value return of map access.
func GetOk(sessionId string, key interface{}) (interface{}, bool) {
	mutex.RLock()
	if _, ok := data[sessionId]; ok {
		value, ok := data[sessionId][key]
		mutex.RUnlock()
		return value, ok
	}
	mutex.RUnlock()
	return nil, false
}

// GetAll returns all stored values for the request as a map. Nil is returned for invalid requests.
func GetAll(sessionId string) map[interface{}]interface{} {
	mutex.RLock()
	if context, ok := data[sessionId]; ok {
		result := make(map[interface{}]interface{}, len(context))
		for k, v := range context {
			result[k] = v
		}
		mutex.RUnlock()
		return result
	}
	mutex.RUnlock()
	return nil
}

// GetAllOk returns all stored values for the request as a map and a boolean value that indicates if
// the request was registered.
func GetAllOk(sessionId string) (map[interface{}]interface{}, bool) {
	mutex.RLock()
	context, ok := data[sessionId]
	result := make(map[interface{}]interface{}, len(context))
	for k, v := range context {
		result[k] = v
	}
	mutex.RUnlock()
	return result, ok
}

// Delete removes a value stored for a given key in a given request.
func Delete(sessionId string, key interface{}) {
	mutex.Lock()
	if data[sessionId] != nil {
		delete(data[sessionId], key)
	}
	mutex.Unlock()
}

// Clear removes all values stored for a given request.
//
// This is usually called by a handler wrapper to clean up request
// variables at the end of a request lifetime. See ClearHandler().
func Clear(sessionId string) {
	mutex.Lock()
	clear(sessionId)
	mutex.Unlock()
}

// clear is Clear without the lock.
func clear(sessionId string) {
	delete(data, sessionId)
	delete(datat, sessionId)
}

// Purge removes request data stored for longer than maxAge, in seconds.
// It returns the amount of requests removed.
//
// If maxAge <= 0, all request data is removed.
//
// This is only used for sanity check: in case context cleaning was not
// properly set some request data can be kept forever, consuming an increasing
// amount of memory. In case this is detected, Purge() must be called
// periodically until the problem is fixed.
func Purge(maxAge int) int {
	mutex.Lock()
	count := 0
	if maxAge <= 0 {
		count = len(data)
		data = make(map[string]map[interface{}]interface{})
		datat = make(map[string]int64)
	} else {
		min := time.Now().Unix() - int64(maxAge)
		for r := range data {
			if datat[r] < min {
				clear(r)
				count++
			}
		}
	}
	mutex.Unlock()
	return count
}
