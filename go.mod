module github.com/projectriff/developer-utils

go 1.13

require (
	github.com/projectriff/stream-client-go v0.5.1-0.20200225224507-836809a469fa
	github.com/spf13/cobra v0.0.6
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
)

replace github.com/projectriff/stream-client-go => github.com/ericbottard/stream-client-go v0.0.0-20200204150506-7cfb9bf48a59
