package bongo

import (
	"gopkg.in/mgo.v2"
	"math"
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

func (r *ResultSet) Next(doc interface{}) bool {

	// Check if the iter has been instantiated yet
	if !r.loadedIter {
		r.Iter = r.Query.Iter()
		r.loadedIter = true
	}

	gotResult := r.Iter.Next(doc)

	if gotResult {

		if hook, ok := doc.(AfterFindHook); ok {
			err := hook.AfterFind(r.Collection)
			if err != nil {
				r.Error = err
				return false
			}
		}

		if newt, ok := doc.(NewTracker); ok {
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

	info := new(PaginationInfo)

	// Get count on a different session to avoid blocking
	sess := r.Collection.Connection.Session.Copy()

	count, err := sess.DB(r.Collection.Database).C(r.Collection.Name).Find(r.Params).Count()
	sess.Close()

	if err != nil {
		return info, err
	}

	// Calculate how many pages
	totalPages := int(math.Ceil(float64(count) / float64(perPage)))

	if page < 1 {
		page = 1
	} else if page > totalPages {
		page = totalPages
	}

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

	return info, nil
}
