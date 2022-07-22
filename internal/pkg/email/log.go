package email

type Log struct {
	ID        string `json:"id" bson:"_id"`
	Service   string `json:"appname" bson:"service"`
	Timestmap int64  `json:"timestamp" bson:"timestamp"`
	EmailID   string `json:"email_id" bson:"email_id"`
	Status    int    `json:"status" bson:"status"`
	Error     string `json:"error" bson:"error"`
}
