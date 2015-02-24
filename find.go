package bongo

import (
	"github.com/maxwellhealth/mgo"
	"log"
	"math"
	"time"
)

type ResultSet struct {
	Query      *mgo.Query
	Iter       *mgo.Iter
	loadedIter bool
	Collection *Collection
	Error      error
	Params     interface{}
}

type PaginationInfo struct {
	Current       int `json:"current"`
	TotalPages    int `json:"totalPages"`
	PerPage       int `json:"perPage"`
	TotalRecords  int `json:"totalRecords"`
	RecordsOnPage int `json:"recordsOnPage"`
}

func (r *ResultSet) Next(mod interface{}) bool {

	// Check if the iter has been instantiated yet
	if !r.loadedIter {
		r.Iter = r.Query.Iter()
		r.loadedIter = true
	}

	gotResult := r.Iter.Next(mod)

	if gotResult {

		if hook, ok := mod.(interface {
			AfterFind(*Collection)
		}); ok {
			hook.AfterFind(r.Collection)
		}

		if newt, ok := mod.(NewTracker); ok {
			newt.SetIsNew(false)
		}
		return true
	}

	err := r.Iter.Err()
	if err != nil {
		r.Error = err
	}

	return false
}

func (r *ResultSet) Free() error {
	if r.loadedIter {
		if err := r.Iter.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Set skip + limit on the current query and generates a PaginationInfo struct with info for your front end
func (r *ResultSet) Paginate(perPage, page int) (*PaginationInfo, error) {
	start := time.Now()
	elapsed := time.Since(start)

	log.Printf("Line 74: %s", elapsed)
	info := new(PaginationInfo)

	// Get count of current query
	// count, err := r.Query.Count()

	sess := r.Collection.Connection.Session.Copy()
	defer sess.Close()

	elapsed = time.Since(start)

	log.Printf("Line 85: %s", elapsed)
	count, err := sess.DB(r.Collection.Connection.Config.Database).C(r.Collection.Name).Find(r.Params).Count()
	// count, err := r.Collection.Collection().Count()

	if err != nil {
		return info, err
	}

	elapsed = time.Since(start)

	log.Printf("Line 95: %s", elapsed)

	// Calculate how many pages
	totalPages := int(math.Ceil(float64(count) / float64(perPage)))

	if page < 1 {
		page = 1
	} else if page > totalPages {
		page = totalPages
	}

	elapsed = time.Since(start)

	log.Printf("Line 108: %s", elapsed)

	skip := (page - 1) * perPage

	r.Query.Skip(skip).Limit(perPage)

	info.TotalPages = totalPages
	info.PerPage = perPage
	info.Current = page
	info.TotalRecords = count

	if info.Current < info.TotalPages {
		info.RecordsOnPage = info.PerPage
	} else {

		info.RecordsOnPage = int(math.Mod(float64(count), float64(perPage)))

		if info.RecordsOnPage == 0 && count > 0 {
			info.RecordsOnPage = perPage
		}

	}

	elapsed = time.Since(start)

	log.Printf("Line 133: %s", elapsed)

	return info, nil
}
