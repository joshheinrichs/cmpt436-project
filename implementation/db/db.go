package db

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
	"time"
)

type User struct {
	ID         bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name       string        `json:"name"`
	Phone      string        `json:"phone"`
	Email      string        `json:"email"`
	IsRealUser bool          `json:"isRealUser"`
	Groups     []Group       `json:"groups" bson:"groups"`
	Contacts   []Contact     `json:"contacts" bson:"contacts"`
	AvatarUrl  string        `json:"avatar_url"`
	Timestamp  time.Time     `json:"time"`
}

type Contact struct {
	ID         bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name       string        `json:"name"`
	Phone      string        `json:"phone"`
	Email      string        `json:"email"`
	IsRealUser bool          `json:"isRealUser`
	Timestamp  time.Time     `json:"time"`
}

type Group struct {
	ID        bson.ObjectId   `json:"id" bson:"_id"`
	GroupName string          `json:"groupName"`
	UserIDs   []bson.ObjectId `json:"users"`
	Expected  []int           `json:"expected"`
	Actual    []int           `json:"actual"`
}

type Comment struct {
	//William changed this to int for testing purposes.
	ID       int    `json:"id" bson:"_id, omitempty"`
	UserName string `json:"username"`
	Content  string `json:"content"`
	//William changed this from time.Time to int for testing purposes.
	Timestamp int64 `json:"time"`
}

type Payment struct {
	ID            bson.ObjectId `json:"id" bson:"_id, omitempty"`
	Payer         string        `json:"payer"`
	Payee         string        `json:"payee"`
	AmountInCents int           `json:"amountInCents"`
	Timestamp     time.Time     `json:"time"`
}

type Purchase struct {
	ID            bson.ObjectId `json:"id" bson:"_id, omitempty"`
	Payer         string        `json:"payer"`
	AmountInCents int           `json:"amountInCents"`
	Timestamp     time.Time     `json:"time"`
}

type Notification struct {
	ID        bson.ObjectId `json:"id" bson:"_id, omitempty"`
	Subject   string        `json:"subject"`
	Content   string        `json:"content"`
	Timestamp time.Time     `json:"time"`
}

var (
	IsDrop     = true
	Session    *mgo.Session
	Collection *mgo.Database
	err        error
)

func ThisPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func ConnectToDB() {

	Session, err = mgo.Dial("127.0.0.1")
	ThisPanic(err)
	Collection = Session.DB("")

}

func Init() *mgo.Collection {

	ConnectToDB()
	ThisPanic(err)

	defer Session.Close()

	Session.SetMode(mgo.Monotonic, true)

	// Drop Database
	if IsDrop {
		err = Session.DB("test").DropDatabase()
		ThisPanic(err)

	}
	c := Session.DB("test").C("User")

	index := mgo.Index{
		Key:        []string{"name", "phone"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(index)

	ThisPanic(err)
	return c
}

//func AddUser(name string, email string, phone string, isRealUser bool) {
//	c := Session.DB("test").C("User")
//	err = c.Insert(&User{Name: name, Phone: phone, IsRealUser: isRealUser, Email: email, Timestamp: time.Now()})
//{}

// func FindUserByID(id bson.ObjectId)

// func GetIDbyEmail(email string)

// func AddGroup(groupName string, id bson.ObjectId)

// func FindGroup(id bson.ObjectId) g *Group

// func AddMemberToGroupByID(server.GroupId bson.ObjectId, userId bson.ObjectId )

// func GetGroupChanges(g Group)

// func RemoveMemberFromGroup(server.GroupId bson.ObjectId, userId bson.ObjectId )

// func DeleteGroup(id bson.ObjectId) b bool

func TestServer(t *testing.T) {

	ConnectToDB()
	ThisPanic(err)

	defer Session.Close()

	Session.SetMode(mgo.Monotonic, true)

	// Drop Database
	if IsDrop {
		err = Session.DB("test").DropDatabase()
		ThisPanic(err)

	}
	c := Session.DB("test").C("User")

	index := mgo.Index{
		Key:        []string{"name", "phone"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(index)

	ThisPanic(err)

	err = c.Insert(&User{Name: "Ale", Phone: "+922", IsRealUser: true, Email: "abc@gmail.com", Timestamp: time.Now()})
	ThisPanic(err)
	err = c.Insert(&User{Name: "Jrock", Phone: "+911", IsRealUser: true, Email: "jcl@gmail.com", Timestamp: time.Now()})
	ThisPanic(err)

	c = Session.DB("test").C("Contact")
	err = c.Insert(&Contact{Name: "Ale", Phone: "+922", IsRealUser: true, Email: "abc@gmail.com", Timestamp: time.Now()})
	ThisPanic(err)

	result := Contact{}
	err = c.Find(bson.M{"name": "Ale"}).One(&result)
	ThisPanic(err)

	fmt.Println("\n")
	fmt.Println(result)
	fmt.Println("\n")

	findJ := User{}
	c = Session.DB("test").C("User")
	err = c.Find(bson.M{"name": "Jrock"}).Select(bson.M{"_id": 1}).One(&findJ)
	fmt.Println(findJ)
	ThisPanic(err)

	fmt.Println("\nHexID of JRock\n")
	fmt.Println(findJ.ID.Hex())
	fmt.Println("\nResult Object\n")
	fmt.Println(result)

	query := bson.M{"_id": bson.ObjectId(findJ.ID)}
	fmt.Println("\nQuery\n")
	fmt.Println(query)
	change := bson.M{"$push": bson.M{"contacts": result}}
	//change2 := bson.M{"$push": bson.M{"contacts": bson.M{"name": result.Name}}}

	fmt.Println("\nUpdate Params\n")
	fmt.Println(change)
	err = c.Update(query, change)
	ThisPanic(err)

	findJ = User{}
	err = c.Find(bson.M{"name": "Jrock"}).One(&findJ)
	ThisPanic(err)

	fmt.Println("\nContacts of JRock\n")
	fmt.Println(findJ.Contacts[0])

	c = Session.DB("test").C("Contact")
	err = c.Insert(&Contact{Name: "Eclo", Phone: "+306", IsRealUser: true, Email: "eclo@gmail.com", Timestamp: time.Now()})
	ThisPanic(err)
	result = Contact{}
	err = c.Find(bson.M{"name": "Eclo"}).One(&result)

	c = Session.DB("test").C("User")
	/*ADD ANOTHER CONTACT*/
	findJ = User{}
	c = Session.DB("test").C("User")
	err = c.Find(bson.M{"name": "Jrock"}).Select(bson.M{"_id": 1}).One(&findJ)
	fmt.Println(findJ)

	ThisPanic(err)

	fmt.Println("\nHexID of JRock\n")
	fmt.Println(findJ.ID.Hex())
	fmt.Println("\nResult Object\n")
	fmt.Println(result)

	query = bson.M{"_id": bson.ObjectId(findJ.ID)}
	fmt.Println("\nQuery\n")
	fmt.Println(query)
	change = bson.M{"$push": bson.M{"contacts": result}}
	//change2 := bson.M{"$push": bson.M{"contacts": bson.M{"name": result.Name}}}

	fmt.Println("\nUpdate Params\n")
	fmt.Println(change)
	err = c.Update(query, change)
	ThisPanic(err)

	findJ = User{}
	err = c.Find(bson.M{"name": "Jrock"}).One(&findJ)
	ThisPanic(err)
	fmt.Println(findJ)
	array := []bson.ObjectId{findJ.ID}

	c = Session.DB("test").C("Group")
	err = c.Insert(&Group{GroupName: "test", UserIDs: array})
	ThisPanic(err)

}
