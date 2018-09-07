// Copyright 2018 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package versioned

import (
	"github.com/cilium/cilium/pkg/lock"
)

// UUID is the UUID for the object that is going to be stored in the map.
type UUID string

// DeepEqualFunc should return true or false if both interfaces `o1` and `o2`
// are considered equal.
type DeepEqualFunc func(o1, o2 interface{}) bool

// ComparableMap is a map that can store Objects that are comparable between
// each other.
type ComparableMap struct {
	Map        map[UUID]Object
	DeepEquals DeepEqualFunc
}

// NewComparableMap returns an initialized map with the equalFunc set as the
// DeepEquals of the map.
func NewComparableMap(equalFunc DeepEqualFunc) *ComparableMap {
	return &ComparableMap{
		Map:        make(map[UUID]Object),
		DeepEquals: equalFunc,
	}
}

// AddEqual maps `uuid` to `obj` if the object to be inserted has a newer
// Version than the one already mapped in the map. Returns false if the object
// inserted does is not mapped yet or if the object has a newer version and
// is not deep equaled than the object already stored.
func (m *ComparableMap) AddEqual(uuid UUID, obj Object) bool {
	oldObj, ok := m.Map[uuid]
	if ok {
		// small performance optimization where we only add
		// an object if the version is newer than the one we have.
		if obj.CompareVersion(oldObj) > 0 {
			m.Map[uuid] = obj
			return m.DeepEquals(oldObj.Data, obj.Data)
		}
		return true
	}
	m.Map[uuid] = obj
	return false
}

// Add maps the uuid to the given obj without any comparison.
func (m *ComparableMap) Add(uuid UUID, obj Object) {
	m.Map[uuid] = obj
}

// Delete deletes the value that maps uuid in the map. Returns true of false
// if the object existed in the map before deletion.
func (m *ComparableMap) Delete(uuid UUID) bool {
	_, exists := m.Map[uuid]
	if exists {
		delete(m.Map, uuid)
	}
	return exists
}

// Get returns the object that maps to the given uuid and returns true or false
// either the object exists or not.
func (m *ComparableMap) Get(uuid UUID) (Object, bool) {
	o, exists := m.Map[uuid]
	return o, exists
}

// SyncComparableMap is a thread-safe wrapper around ComparableMap.
type SyncComparableMap struct {
	mutex *lock.RWMutex
	cm    *ComparableMap
}

// NewSyncComparableMap returns a thread-safe ComparableMap.
func NewSyncComparableMap(def DeepEqualFunc) *SyncComparableMap {
	return &SyncComparableMap{
		cm: NewComparableMap(def),
	}
}

// Add maps the uuid to the given obj without any comparison.
func (sm *SyncComparableMap) Add(uuid UUID, obj Object) {
	sm.mutex.Lock()
	sm.cm.Add(uuid, obj)
	sm.mutex.Unlock()
}

// AddEqual maps `uuid` to `obj` if the object to be inserted has a newer
// Version than the one already mapped in the map. Returns false if the object
// inserted does is not mapped yet or if the object has a newer version and
// is not deep equaled than the object already stored.
func (sm *SyncComparableMap) AddEqual(uuid UUID, obj Object) bool {
	sm.mutex.Lock()
	added := sm.cm.AddEqual(uuid, obj)
	sm.mutex.Unlock()
	return added
}

// Delete deletes the value that maps uuid in the map. Returns true of false
// if the object existed in the map before deletion.
func (sm *SyncComparableMap) Delete(uuid UUID) bool {
	sm.mutex.Lock()
	exists := sm.cm.Delete(uuid)
	sm.mutex.Unlock()
	return exists
}

// Get returns the object that maps to the given uuid and returns true or false
// either the object exists or not.
func (sm *SyncComparableMap) Get(uuid UUID) (Object, bool) {
	sm.mutex.RLock()
	v, e := sm.cm.Get(uuid)
	sm.mutex.RUnlock()
	return v, e
}

// DoLocked is a thread-safe function that can be used to perform multiple
// operations in the map atomically.
// Parameters:
//  * iterate: if not nil, the ComparableMap will iterate over all keys and call
//    this function for every key-value pair. Both key and value should only
//    be used for read.
//  * replace: if not nil, replace is called with the internal ComparableMap,
//    the returned ComparableMap will be set as the new internal ComparableMap.
//    If an error is returned, the replace operation won't take place.
// In case both `iterate` and `replace` are provided, `iterate` is executed
// first and `replace` is executed afterwards.
func (sm *SyncComparableMap) DoLocked(iterate func(key UUID, value Object), replace func(old *ComparableMap) (*ComparableMap, error)) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	if iterate != nil {
		for k, v := range sm.cm.Map {
			iterate(k, v)
		}
	}
	if replace != nil {
		newMap, err := replace(sm.cm)
		if err != nil {
			return err
		}
		sm.cm = newMap
	}
	return nil
}
