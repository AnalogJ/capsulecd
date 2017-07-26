package pipeline

type ScmReleaseAsset struct { //mapstructure is used to deserialize by Config.
	LocalPath    string `mapstructure:"local_path"`
	ArtifactName string `mapstructure:"artifact_name"`
	ContentType  string `mapstructure:"content_type"`
}
