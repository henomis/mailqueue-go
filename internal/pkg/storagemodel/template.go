package storagemodel

type Template struct {
	ID       string `bson:"_id"`
	Name     string `bson:"name"`
	Template string `bson:"template"`
}
