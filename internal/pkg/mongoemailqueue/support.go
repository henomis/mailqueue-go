package mongoemailqueue

import "fmt"

func validateMongoQueueOptions(mongoQueueOptions *MongoEmailQueueOptions) error {

	if len(mongoQueueOptions.Endpoint) == 0 {
		return fmt.Errorf("invalid endpoint")
	}

	if len(mongoQueueOptions.Database) == 0 {
		return fmt.Errorf("invalid database name")
	}

	if len(mongoQueueOptions.Collection) == 0 {
		return fmt.Errorf("invalid collection name")
	}

	if mongoQueueOptions.CappedSize == 0 {
		return fmt.Errorf("invalid capped size")
	}

	return nil
}
