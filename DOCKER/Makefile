build:
	@sh -c "'$(CURDIR)/build.sh'"

push:
	@sh -c "'$(CURDIR)/push.sh'"

build_develop:
	docker build -t "teragrid/teragrid:develop" -f Dockerfile.develop .

push_develop:
	docker push "teragrid/teragrid:develop"

.PHONY: build build_develop push push_develop
