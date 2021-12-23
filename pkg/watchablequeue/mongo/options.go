package mongowatchablequeue

type MongoWatchableQueueOptions struct {
	MongoEndpoint            string
	MongoDatabase            string
	MongoCollection          string
	MongoCappedSize          int64
	MongoDocumentFilterQuery string
	MongoUpdateOnCommitQuery string
	MongoSetStatusQuery      string
}
