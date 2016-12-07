package mock

//go:generate mockgen -package gcloud -destination gcloud/api.go github.com/docker/infrakit/pkg/example/instance/gce/gcloud GCloud
