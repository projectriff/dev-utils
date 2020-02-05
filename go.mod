module github.com/projectriff/developer-utils

go 1.13

require (
	github.com/projectriff/stream-client-go v0.0.0-20200203203246-b0f81b40a96f
	github.com/spf13/cobra v0.0.5
	k8s.io/apimachinery v0.16.6
	k8s.io/client-go v0.16.6
)

replace github.com/projectriff/stream-client-go => github.com/ericbottard/stream-client-go v0.0.0-20200204150506-7cfb9bf48a59
