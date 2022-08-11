package mongoemaillog

import "fmt"

func validateMongoEmailLogOptions(mongoEmailLogOptions *MongoEmailLogOptions) error {

	if len(mongoEmailLogOptions.Endpoint) == 0 {
		return fmt.Errorf("invalid endpoint")
	}

	if len(mongoEmailLogOptions.Database) == 0 {
		return fmt.Errorf("invalid database name")
	}

	if len(mongoEmailLogOptions.Collection) == 0 {
		return fmt.Errorf("invalid collection name")
	}

	return nil
}
