tf-test:
	TF_ACC=1 go test -v -timeout 120m -coverprofile cover.out -parallel 4 ./internal/...

docs:
	tfplugindocs generate -rendered-provider-name "Discord" --provider-name "discord"

tf-provider-lint:
	tfproviderlintx -AT003=false ./internal/...