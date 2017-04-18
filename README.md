# What's Bongo?
We couldn't find a good ODM for MongoDB written in Go, so we made one. Bongo is a wrapper for mgo (https://github.com/go-mgo/mgo) that adds ODM, hooks, validation, and cascade support to its raw Mongo functions.

Bongo is tested using the fantasic GoConvey (https://github.com/smartystreets/goconvey)

[![Build Status](https://travis-ci.org/go-bongo/bongo.svg)](https://travis-ci.org/go-bongo/bongo)

[![Coverage Status](https://coveralls.io/repos/go-bongo/bongo/badge.svg)](https://coveralls.io/r/go-bongo/bongo)

# Stablity

Since we're not yet at a major release, some things in the API might change. Here's a list:

* Save - stable
* Find/FindOne/FindById - stable
* Delete - stable
* Save/Delete/Find/Validation hooks - stable
* Cascade - unstable (might need a refactor)
* Change Tracking - stable
* Validation methods - stable

# Usage

## Basic Usage
### Import the Library
`go get github.com/go-bongo/bongo`

`import "github.com/go-bongo/bongo"`

And install dependencies:

`cd $GOHOME/src/github.com/go-bongo/bongo && go get .`

### Connect to a Database

Create a new `bongo.Config` instance:

```go
config := &bongo.Config{
	ConnectionString: "localhost",
	Database:         "bongotest",
}
```

Then just create a new instance of `bongo.Connection`, and make sure to handle any connection errors:

```go
connection, err := bongo.Connect(config)

if err != nil {
	log.Fatal(err)
}
```

If you need to, you can access the raw `mgo` session with `connection.Session`

### Create a Document

Any struct can be used as a document as long as it satisfies the `Document` interface (`SetId(bson.ObjectId)`, `GetId() bson.ObjectId`). We recommend that you use the `DocumentBase` provided with Bongo, which implements that interface as well as the `NewTracker`, `TimeCreatedTracker` and `TimeModifiedTracker` interfaces (to keep track of new/existing documents and created/modified timestamps). If you use the `DocumentBase` or something similar, make sure you use `bson:",inline"` otherwise you will get nested behavior when the data goes to your database.

For example:

```go
type Person struct {
	bongo.DocumentBase `bson:",inline"`
	FirstName string
	LastName string
	Gender string
}
```

You can use child structs as well.

```go
type Person struct {
	bongo.DocumentBase `bson:",inline"`
	FirstName string
	LastName string
	Gender string
	HomeAddress struct {
		Street string
		Suite string
		City string
		State string
		Zip string
	}
}
```

#### Hooks

You can add special methods to your document type that will automatically get called by bongo during certain actions. Hooks get passed the current `*bongo.Collection` so you can avoid having to couple them with your actual database layer. Currently available hooks are:

* `func (s *ModelStruct) Validate(*bongo.Collection) []error` (returns a slice of errors - if it is empty then it is assumed that validation succeeded)
* `func (s *ModelStruct) BeforeSave(*bongo.Collection) error`
* `func (s *ModelStruct) AfterSave(*bongo.Collection) error`
* `func (s *ModelStruct) BeforeDelete(*bongo.Collection) error`
* `func (s *ModelStruct) AfterDelete(*bongo.Collection) error`
* `func (s *ModelStruct) AfterFind(*bongo.Collection) error`

### Saving Models

Just call `save` on a collection instance.

```go
myPerson := &Person{
	FirstName:"Testy",
	LastName:"McGee",
	Gender:"male",
}
err := connection.Collection("people").Save(myPerson)
```

Now you'll have a new document in the `people` collection. If there is an error, you can check if it is a validation error using a type assertion:

```go
if vErr, ok := err.(*bongo.ValidationError); ok {
	fmt.Println("Validation errors are:", vErr.Errors)
} else {
	fmt.Println("Got a real error:", err.Error())
}
```

### Deleting Documents

There are three ways to delete a document.

#### DeleteDocument
Same thing as `Save` - just call `DeleteDocument` on the collection and pass the document instance.
```go
err := connection.Collection("people").DeleteDocument(person)
```

This *will* run the `BeforeDelete` and `AfterDelete` hooks, if applicable.

#### DeleteOne
This just delegates to `mgo.Collection.Remove`. It will *not* run the `BeforeDelete` and `AfterDelete` hooks.

```go
err := connection.Collection("people").DeleteOne(bson.M{"FirstName":"Testy"})
```

#### Delete
This delegates to `mgo.Collection.RemoveAll`. It will *not* run the `BeforeDelete` and `AfterDelete` hooks.
```go
changeInfo, err := connection.Collection("people").Delete(bson.M{"FirstName":"Testy"})
fmt.Printf("Deleted %d documents", changeInfo.Removed)
```


### Find by ID

```go
person := &Person{}
err := connection.Collection("people").FindById(bson.ObjectIdHex(StringId), person)
```

The error returned can be a `DocumentNotFoundError` or a more low-level MongoDB error. To check, use a type assertion:

```go
if dnfError, ok := err.(*bongo.DocumentNotFoundError); ok {
	fmt.Println("document not found")
} else {
	fmt.Println("real error " + err.Error())
}
```

### Find

Finds will return an instance of `ResultSet`, which you can then optionally `Paginate` and iterate through to get all results.

```go

// *bongo.ResultSet
results := connection.Collection("people").Find(bson.M{"firstName":"Bob"})

person := &Person{}

count := 0

for results.Next(person) {
	fmt.Println(person.FirstName)
}
```

To paginate, you can run `Paginate(perPage int, currentPage int)` on the result of `connection.Find()`. That will return an instance of `bongo.PaginationInfo`, with properties like `TotalRecords`, `RecordsOnPage`, etc.

To use additional functions like `sort`, `skip`, `limit`, etc, you can access the underlying mgo `Query` via `ResultSet.Query`.

### Find One
Same as find, but it will populate the reference of the struct you provide as the second argument.


```go

person := &Person{}

err := connection.Collection("people").FindOne(bson.M{"firstName":"Bob"}, person)

if err != nil {
	fmt.Println(err.Error())
} else {
	fmt.Println("Found user:", person.FirstName)
}
```

## Change Tracking
If your model struct implements the `Trackable` interface, it will automatically track changes to your model so you can compare the current values with the original. For example:

```go
type MyModel struct {
	bongo.DocumentBase `bson:",inline"`
	StringVal string
	diffTracker *bongo.DiffTracker
}

// Easy way to lazy load a diff tracker
func (m *MyModel) GetDiffTracker() *DiffTracker {
	if m.diffTracker == nil {
		m.diffTracker = bongo.NewDiffTracker(m)
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

### Diff-tracking Session
If you are going to be checking more than one field, you should instantiate a new `DiffTrackingSession` with `diffTracker.NewSession(useBsonTags bool)`. This will load the changed fields into the session. Otherwise with each call to `diffTracker.Modified()`, it will have to recalculate the changed fields.


## Cascade Save/Delete
Bongo supports cascading portions of documents to related documents and the subsequent cleanup upon deletion. For example, if you have a `Team` collection, and each team has an array of `Players`, you can cascade a player's first name and last name to his or her `team.Players` array on save, and remove that element in the array if you delete the player.

To use this feature, your struct needs to have an exported method called `GetCascade`, which returns an array of `*bongo.CascadeConfig`. Additionally, if you want to make use of the `OldQuery` property to remove references from previously related documents, you should probably alsotimplement the `DiffTracker` on your model struct (see above).

You can also leave `ThroughProp` blank, in which case the properties of the document will be cascaded directly onto the related document. This is useful when you want to cascade `ObjectId` properties or other references, but it is important that you keep in mind that these properties will be nullified on the related document when the main doc is deleted or changes references.

Also note that like the above hooks, the `GetCascade` method will be passed the instance of the `bongo.Collection` so you can keep your models decoupled from your database layer.

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

	// Properties that will be cascaded/deleted. Can (should) be in dot notation for nested properties. This is used to nullify properties when there is an OldQuery or if the document is deleted.
	Properties []string

	// The actual data that will be cascade
	Data interface{}
}
```

### Example
```go
type ChildRef struct {
	Id bson.ObjectId `bson:"_id" json:"_id"`
	Name string
}
func (c *Child) GetCascade(collection *bongo.Collection) []*bongo.CascadeConfig {
	connection := collection.Connection
	rel := &ChildRef {
		Id:c.Id,
		Name:c.Name,
	}
	cascadeSingle := &bongo.CascadeConfig{
		Collection:  connection.Collection("parents").Collection(),
		Properties:  []string{"name"},
		Data:rel,
		ThroughProp: "child",
		RelType:     bongo.REL_ONE,
		Query: bson.M{
			"_id": c.ParentId,
		},
	}

	cascadeMulti := &bongo.CascadeConfig{
		Collection:  connection.Collection("parents").Collection(),
		Properties:  []string{"name"},
		Data:rel,
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

1. When you save a child, it will populate its parent's (defined by `cascadeSingle.Query`) `child` property with an object, consisting of one key/value pair (`name`)

2. When you save a child, it will also modify its parent's (defined by `cascadeMulti.Query`) `children` array, either modifying or pushing to the array of key/value pairs, also with just `name`.

3. When you delete a child, it will use `cascadeSingle.OldQuery` to remove the reference from its previous `parent.child`

4. When you delete a child, it will also use `cascadeMulti.OldQuery` to remove the reference from its previous `parent.children`

Note that the `ThroughProp` must be the actual field name in the database (bson tag), not the property name on the struct. If there is no `ThroughProp`, the data will be cascaded directly onto the root of the document.
