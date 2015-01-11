# What's Bongo?
We couldn't find a good ODM for MongoDB written in Go, so we made one. Bongo is a wrapper for mgo (https://github.com/go-mgo/mgo) that adds ODM, hooks, validation, cascade support, and HIPAA-compliant encryption to its standard Mongo functions.

# Usage

## Basic Usage
### Import the Library
`go get github.com/maxwellhealth/bongo`

`import "github.com/maxwellhealth/bongo"`

And install dependencies:

`cd $GOHOME/src/github.com/maxwellhealth/bongo && go get .`

### Connect to a Database

Create a new `bongo.Config` instance:

```go
config := &bongo.Config{
	ConnectionString: "localhost",
	Database:         "bongotest",
	EncryptionKey:    []byte("MyEncryptionKey"),
}
```

Yep! Bongo has built-in support for encrypted fields (for HIPAA compliance) and even encryption keys per collection (use the `EncryptionKeyPerCollection map[string]string`).

Then just create a new instance of `bongo.Connection`:

```go
connection := bongo.Connect(config)
```

If you need to, you can access the raw `mgo` session with `connection.Session`

### Create a Model

Any struct can be used as a model as long as it has an Id property with type `bson.ObjectId` (from `mgo/bson`). `bson` tags are passed through to mgo. You can specify a field as being encrypted using `bongo:"encrypted"`

For example:

```go

type Person struct {
	Id bson.ObjectId `bson:"_id"`
	FirstName string `bongo:"encrypted" bson:"firstName"`
	LastName string `bongo:"encrypted" bson:"lastName"`
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
	Id bson.ObjectId `bson:"_id"`
	FirstName string `bongo:"encrypted" bson:"firstName"`
	LastName string `bongo:"encrypted" bson:"lastName"`
	Gender string
	HomeAddress Address `bongo:"encrypted" bson:"homeAddress"`
}
```

### Indexing

You can use tags to ensure indeces on your collections. The mere presence of an `index` tag will cause Bongo to ensure an index on that field when your model is registered. If you also have the "unique" tag, it will be a unique index.

```go
type Person struct {
	Id bson.ObjectId `bson:"_id"`
	FirstName string `bongo:"encrypted" bson:"firstName"`
	LastName string `bongo:"encrypted" bson:"lastName"`
	Gender string `bongo:"index"`
	Email string `bongo:"index,unique"`
}
```

To register your model, you should do the following at boot time. This will ensure the indeces defined in `Person` will be present in the `"people"` collection. If you leave the second argument as a blank string, it will interpret the collection name from the name of the struct (in this case getting "person")

```go
connection.Register(&Person{}, "people")
```

#### Hooks

You can add special methods to your struct that will automatically get called by bongo during certain actions. Hooks get passed the current `*bongo.Collection` so you can avoid having to couple them with your actual database layer. Currently available hooks are:

* `func (s *ModelStruct) Validate(*bongo.Collection) []string` (returns a slice of errors)
* `func (s *ModelStruct) BeforeSave(*bongo.Collection)`
* `func (s *ModelStruct) BeforeCreate(*bongo.Collection)`
* `func (s *ModelStruct) BeforeUpdate(*bongo.Collection)`
* `func (s *ModelStruct) AfterSave(*bongo.Collection)`
* `func (s *ModelStruct) AfterCreate(*bongo.Collection)`
* `func (s *ModelStruct) AfterUpdate(*bongo.Collection)`
* `func (s *ModelStruct) AfterFind(*bongo.Collection)`

The create/update hooks run immediately before the save hooks.
	
#### Validation

Use the `Validate()` hook to validate your model. If you return a slice with at least one element, the `Save()` method will fail. Bongo comes with some built-in validation methods:

* `func bongo.ValidateRequired(val interface{}) bool` - makes sure the provided val is not equal to its type's zero-value
* `func bongo.ValidateMongoIdRef(val interface{}, collection *bongo.Collection) bool` - makes sure the provided val (`bson.ObjectId`) references a document in the provided collection
* `func bongo.ValidateInclusionIn(value string, options []string) bool` - make sure the provided `string` val matches an element in the given options

You can obviously use your own validation as long as you add elements to the returned `[]string`

### Saving Models

Bongo can intelligently guess the name of the collection using the name of the struct you pass. (e.g. "FooBar" would go in as "foo_bar"). If you're OK with that, you can save directly via your connection:

```go
myPerson := &Person{
	FirstName:"Testy",
	LastName:"McGee",
	Gender:"male",
}
saveResult := connection.Save(myPerson)
```

You will now have a new document in the `person` collection.

To insert this into a collection called "people", you can do the following:

```go
myPerson := &Person{
	FirstName:"Testy",
	LastName:"McGee",
	Gender:"male",
}
saveResult := connection.Collection("people").Save(myPerson)
```

Now you'll have a new document in the `people` collection.

### Deleting Models

Same deal as save.

To delete from the "person" collection (assuming person is a full struct with a valid Id property):

```go
err := connection.Delete(person)
```

Or from the "people" collection (same assumption):
```go
err := connection.Collection("people").Delete(person)
```


### Find by ID

Same thing applies re: collection name. This will look in "person" and populate the reference of `person`:

```go
import "labix.org/v2/mgo/bson"

...

person := new(Person)

err := connection.FindById(bson.ObjectIdHex(StringId), person)
```

And this will look in "people":

```go
import "labix.org/v2/mgo/bson"

...

person := new(Person)

err := connection.Collection("people").FindById(bson.ObjectIdHex(StringId), person)
```

### Find

Find's a bit different - it's not a direct operation on a model reference so you can either call it directly on the `bongo.Connection`, passing either a sample struct or the collection name as the second argument so it knows which collection look in. You can also call `Collection.Find`, in which case you will only have to pass one argument (the query).

```go

// *bongo.ResultSet
results := connection.Find(bson.M{"firstName":"Bob"}, "people")

// OR: connection.Collection("people").Find(bson.M{"firstName":"Bob"})

person := new(Person)

count := 0

for results.Next(person) {
	fmt.Println(person.FirstName)
}
```

You can also pass a sample reference as the second argument instead of a string. This will look in the "person" collection instead of "people":

```go
results := connection.Find(nil, &Person{})
```

To paginate, you can run `Paginate(perPage int, currentPage int)` on the result of `connection.Find()`.

To use additional functions like `sort`, you can access the underlying mgo `Query` via `ResultSet.Query`.

### Find One
Same as find, but it will populate the reference of the struct you provide as the second argument. If there is no document found, you will get an error:


```go
import (
	"labix.org/v2/mgo/bson"
	"fmt"
)

...

person := new(Person)

err := connection.FindOne(bson.M{"firstName":"Bob"}, person)

// Or connection.Collection("people").FindOne(bson.M{"firstName":"Bob"}) if you want to search the "people" collection

if err != nil {
	fmt.Println(err.Error())
} else {
	fmt.Println("Found user:", person.firstName)
}
```

## Change Tracking
If your model struct implements the `Trackable` interface, it will automatically track changes to your model so you can compare the current values with the original. For example:

```go
type MyModel struct {
	Id bson.ObjectId `bson:"_id"`
	StringVal string
	diffTracker *bongo.DiffTracker
}

// Easy way to lazy load a diff tracker
func (m *MyModel) GetDiffTracker() *DiffTracker {
	v := reflect.ValueOf(m.diffTracker)
	if !v.IsValid() || v.IsNil() {
		m.diffTracker = NewDiffTracker(m)
	}

	return m.diffTracker
}

myModel := &MyModel{}
```

Use as follows:

### Check if a field has been modified
```go
// Store the current state for comparison
myModel.GetDiffTracker().Reset()

// Change a property...
myModel.StringVal = "foo"

// We know it's been instantiated so no need to use GetDiffTracker()
fmt.Println(myModel.diffTracker.Modified("StringVal")) // true
myModel.diffTracker.Reset()
fmt.Println(myModel.diffTracker.Modified("StringVal")) // false
```

### Get all modified fields
```go
myModel.StringVal = "foo"
// Store the current state for comparison
myModel.GetDiffTracker().Reset()

isNew, modifiedFields := myModel.GetModified()

fmt.Println(isNew, modifiedFields) // false, ["StringVal"]
myModel.diffTracker.Reset()

isNew, modifiedFields = myModel.GetModified()
fmt.Println(isNew, modifiedFields) // false, []
```


## Cascade Save/Delete
Bongo supports cascading portions of documents to related documents and the subsequent cleanup upon deletion. For example, if you have a `Team` collection, and each team has an array of `Players`, you can cascade a player's first name and last name to his or her `team.Players` array on save, and remove that element in the array if you delete the player.

To use this feature, your struct needs to have an exported method called `GetCascade`, which returns an array of `*bongo.CascadeConfig`. Additionally, if you want to make use of the `OldQuery` property to remove references from previously related documents, you should probably alsotimplement the `DiffTracker` on your model struct (see above). 

On the struct properties that are cascaded from related documents, you need to tell Mongo not to save them, and how to decrypt them. (The related collection could have a different encryption key). To do this, use the `cascadedFrom={collectionName}` bongo tag, like so `bongo:"cascadedFrom=children"`. This will tell Bongo not to save those fields when you save your model (since they are supposed to be populated by the related documents), and also to decrypt those fields using the encryption key for the "children" collection, rather than the main model's collection.

You can also leave `ThroughProp` blank, in which case the properties of the document will be cascaded directly onto the related document. This is useful when you want to cascade `ObjectId` properties or other references, but it is important that you keep in mind that (a) these properties will be nullified on the related document when the main doc is deleted or changes references, and (b) they will fail decryption if you have encryption keys per collection, because currently there is no way to designate that property is cascaded from another collection unless it is a struct or slice of structs.

### Casade Configuration
```go
type CascadeConfig struct {
	// The collection to cascade to
	Collection *mgo.Collection

	// The relation type (does the target doc have an array of these docs [REL_MANY] or just reference a single doc [REL_ONE])
	RelType int

	// The property on the related doc to populate
	ThroughProp string

	// The query to find related docs
	Query bson.M

	// The data that constructs the query may have changed - this is to remove self from previous relations
	OldQuery bson.M

	// Data to cascade. Can be in dot notation
	Properties []string
}
```

### Example
```go
func (c *Child) GetCascade() []*bongo.CascadeConfig {

	cascadeSingle := &bongo.CascadeConfig{
		Collection:  connection.Collection("parents").Collection(),
		Properties:  []string{"name"},
		ThroughProp: "child",
		RelType:     bongo.REL_ONE,
		Query: bson.M{
			"_id": c.ParentId,
		},
	}

	cascadeMulti := &bongo.CascadeConfig{
		Collection:  connection.Collection("parents").Collection(),
		Properties:  []string{"name"},
		ThroughProp: "children",
		RelType:     bongo.REL_MANY,
		Query: bson.M{
			"_id": c.ParentId,
		},
	}

	if c.DiffTracker.Modified("ParentId") {

		origId, _ := c.DiffTracker.GetOriginalValue("ParentId")
		if origId != nil {
			oldQuery := bson.M{
				"_id": origId,
			}
			cascadeSingle.OldQuery = oldQuery
			cascadeMulti.OldQuery = oldQuery
		}

	}

	return []*bongo.CascadeConfig{cascadeSingle, cascadeMulti}
}
```

This does the following:

1) When you save a child, it will populate its parent's (defined by `cascadeSingle.Query`) `child` property with an object, consisting of one key/value pair (`name`)

2) When you save a child, it will also modify its parent's (defined by `cascadeMulti.Query`) `children` array, either modifying or pushing to the array of key/value pairs, also with just `name`.

3) When you delete a child, it will use `cascadeSingle.OldQuery` to remove the reference from its previous `parent.child`

4) When you delete a child, it will also use `cascadeMulti.OldQuery` to remove the reference from its previous `parent.children`

Note that the `ThroughProp` must be the actual field name in the database, not the property name on the struct.

# Credits
* Mongoose for inspiration
* Mitchell Hashimoto for his mapstructure repo (https://github.com/mitchellh/mapstructure), the codec for which I shamelessly stole and modified instead of making my own

