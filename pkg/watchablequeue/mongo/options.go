package mongowatchablequeue

type MongoWatchableQueueOptions struct {
	MongoEndpoint       string
	MongoDatabase       string
	MongoCollection     string
	MongoCappedSize     int64
	MongoDocumentFilter string
	MongoUpdateOnCommit string
}
