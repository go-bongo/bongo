package bongo

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

type DocumentBase struct {
	Id       bson.ObjectId `bson:"_id,omitempty" json:"_id"`
	Created  time.Time     `bson:"_created" json:"_created"`
	Modified time.Time     `bson:"_modified" json:"_modified"`

	// We want this to default to false without any work. So this will be the opposite of isNew. We want it to be new unless set to existing
	exists bool
}

// Satisfy the new tracker interface
func (d *DocumentBase) SetIsNew(isNew bool) {
	d.exists = !isNew
}

// Is the document new
func (d *DocumentBase) IsNew() bool {
	return !d.exists
}

// Satisfy the document interface
func (d *DocumentBase) GetId() bson.ObjectId {
	return d.Id
}

// Sets the ID for the document
func (d *DocumentBase) SetId(id bson.ObjectId) {
	d.Id = id
}

// Set's the created date
func (d *DocumentBase) SetCreated(t time.Time) {
	d.Created = t
}

// Get the created date
func (d *DocumentBase) GetCreated() time.Time {
	return d.Created
}

// Sets the modified date
func (d *DocumentBase) SetModified(t time.Time) {
	d.Modified = t
}

// Get's the modified date
func (d *DocumentBase) GetModified() time.Time {
	return d.Modified
}
