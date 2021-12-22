package mongowatchablequeue

import "fmt"

type MongoElement struct {
	ID    string      `bson:"_id"`
	Value interface{} `bson:"value"`
}

func (m *MongoElement) String() string {
	return fmt.Sprintf("ID: %s, Value: %+v", m.ID, m.Value)
}

func validateMongoElement(element interface{}) (*MongoElement, error) {

	mongoElement, ok := element.(*MongoElement)
	if !ok {
		return nil, fmt.Errorf("invalid element")
	}

	return mongoElement, nil
}
