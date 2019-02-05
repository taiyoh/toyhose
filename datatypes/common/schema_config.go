package common

// https://docs.aws.amazon.com/ja_jp/firehose/latest/APIReference/API_SchemaConfiguration.html

type SchemaConf struct {
	CatalogID    *string `json:"CatalogId"`
	DatabaseName *string `json:"DatabaseName"`
	Region       *string `json:"Region"`
	RoleARN      *string `json:"RoleARN"`
	TableName    *string `json:"TableName"`
	VersionID    *string `json:"VersionId"`
}
