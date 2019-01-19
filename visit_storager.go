package gocrawler

import "sync"

// VisitStorager - interface of detecting first visit to site
type VisitStorager interface {
	IsFisrtVisit(url string) bool
}

// VisitStorageMapMutex - detector first visited on map with mutex
type VisitStorageMapMutex struct {
	mx           sync.Mutex
	visitedSites map[string]struct{}
}

// NewVisitedStorageMapMutex - constructor
func NewVisitedStorageMapMutex() *VisitStorageMapMutex {
	return &VisitStorageMapMutex{
		visitedSites: make(map[string]struct{}),
	}
}

// IsFisrtVisit - check if visit is first
func (v *VisitStorageMapMutex) IsFisrtVisit(url string) bool {
	v.mx.Lock()
	_, ok := v.visitedSites[url]
	if !ok {
		v.visitedSites[url] = struct{}{}
	}
	v.mx.Unlock()
	return !ok
}
