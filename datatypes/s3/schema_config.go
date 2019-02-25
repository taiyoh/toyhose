package s3

// https://docs.aws.amazon.com/ja_jp/firehose/latest/APIReference/API_SchemaConfiguration.html

// SchemaConf specifies the schema to which you want Kinesis Data Firehose to configure your data before it writes it to Amazon S3.
type SchemaConf struct {
	CatalogID    *string `json:"CatalogId"`
	DatabaseName *string `json:"DatabaseName"`
	Region       *string `json:"Region"`
	RoleARN      *string `json:"RoleARN"`
	TableName    *string `json:"TableName"`
	VersionID    *string `json:"VersionId"`
}
