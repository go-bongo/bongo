# What's Bongo?
We couldn't find a good ODM for MongoDB written in Go, so we made one. Bongo is a wrapper for mgo (https://github.com/go-mgo/mgo) that adds ODM and hook functionality to its standard Mongo functions. It's pretty basic for now, but we are adding features constantly. 

# Usage

## Import the Library
`go get github.com/maxwellhealth/bongo`

`import "github.com/maxwellhealth/bongo"`

## Connect to a Database

Create a new `bongo.Config` instance:

```go
config := &bongo.Config{
	ConnectionString: "localhost",
	Database:         "bongotest",
	EncryptionKey:    "MyEncryptionKey",
}
```

Yep! Bongo has built-in support for encrypted fields (for HIPAA compliance) and even encryption keys per collection (use the `EncryptionKeyPerCollection map[string]string`).

Then just create a new instance of `bongo.Connection`:

```go
connection := bongo.Connect(config)
```

If you need to, you can access the raw `mgo` session with `connection.Session`

## Create a Model

Any struct can be used as a model as long as it has an Id property with type `bson.ObjectId` (from `mgo/bson`). `bson` tags are passed through to mgo. You can specify a field as being encrypted using `encrypted:"true"`

For example:

```go

type Person struct {
	FirstName string `encrypted:"true" bson:"firstName"`
	LastName string `encrypted:"true" bson:"lastName"`
	Gender string
}
```

You can use child structs as well. If encrypted, they will be inserted into the database as one field (encrypted json-encoded string).

```go
type Address struct {
	Street string
	Suite string
	City string
	State string
	Zip string
}

type Person struct {
	FirstName string `encrypted:"true" bson:"firstName"`
	LastName string `encrypted:"true" bson:"lastName"`
	Gender string
	HomeAddress Address `encrypted:"true" bson:"homeAddress"`
}
```

### Hooks

You can add special methods to your struct that will automatically get called by bongo during certain actions. Currently available hooks are:

* `func (s *ModelStruct) Validate() []string` (returns a slice of errors)
* `func (s *ModelStruct) BeforeSave()`
* `func (s *ModelStruct) BeforeCreate()`
* `func (s *ModelStruct) BeforeUpdate()`
* `func (s *ModelStruct) AfterFind()`
	
### Validation

Use the `Validate()` hook to validate your model. If you return a slice with at least one element, the `Save()` method will fail. Bongo comes with some built-in validation methods:

* `func bongo.ValidateRequired(val interface{}) bool` - makes sure the provided val is not equal to its type's zero-value
* `func bongo.ValidateMongoIdRef(val interface{}, collection *bongo.Collection) bool` - makes sure the provided val (`bson.ObjectId`) references a document in the provided collection
* `func bongo.ValidateInclusionIn(value string, options []string) bool` - make sure the provided `string` val matches an element in the given options

You can obviously use your own validation as long as you add elements to the returned `[]string`

## Saving Models

Coming soon... (check the tests in the meantime)
