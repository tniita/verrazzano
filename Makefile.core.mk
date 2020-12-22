
#GOARCH_LOCAL := $(TARGET_ARCH)
#GOOS_LOCAL := $(TARGET_OS)
# Force to make Linux builds
GOARCH_LOCAL := amd64
GOOS_LOCAL := linux

DOCKER_BUILD_VARIANTS ?= default

RELEASE_LDFLAGS='-extldflags -static -s -w'

DOCKER_IMAGE_TAG ?= local-$(shell git rev-parse --short HEAD)-1

STANDARD_BINARIES:=./operator \
	./application-operator

DOCKER_TARGETS ?= verrazzano-platform-operator verrazzano-application-operator

# create a DOCKER_PUSH_TARGETS that's each of DOCKER_TARGETS with a push. prefix
DOCKER_PUSH_TARGETS:=
$(foreach TGT,$(DOCKER_TARGETS),$(eval DOCKER_PUSH_TARGETS+=push.$(TGT)))

.PHONY: clean
clean:
	rm -rf ${TARGET_OUT}

#build: depend ## Builds all go binaries.
.PHONY: build
build:
	@mkdir -p $(TARGET_OUT)
	GOOS=$(GOOS_LOCAL) GOARCH=$(GOARCH_LOCAL) LDFLAGS=$(RELEASE_LDFLAGS) common/scripts/gobuild.sh $(TARGET_OUT)/ $(STANDARD_BINARIES)

# This target will package all docker images used in test and release, without re-building
# go binaries. It is intended for CI/CD systems where the build is done in separate job.
.PHONY: docker.all
docker.all: build $(DOCKER_TARGETS)

#.PHONY: operator-docker-build
.PHONY: verrazzano-platform-operator
verrazzano-platform-operator:
	$(shell cp operator/Dockerfile.platform ${TARGET_OUT})
	$(shell cp -r operator/scripts ${TARGET_OUT})
	$(shell cp -r operator/config ${TARGET_OUT})
	cd ${TARGET_OUT} && docker build --pull -f Dockerfile.platform \
		-t $(@):${DOCKER_IMAGE_TAG} .

#.PHONY: application-operator-docker-build
.PHONY: verrazzano-application-operator
verrazzano-application-operator:
	$(shell cp application-operator/Dockerfile.application ${TARGET_OUT})
	cd ${TARGET_OUT} && docker build --pull -f Dockerfile.application \
		-t $(@):${DOCKER_IMAGE_TAG} .

# Will build and push docker images.
.PHONY: docker.push
docker.push: $(DOCKER_PUSH_TARGETS)
$(foreach TGT,$(DOCKER_TARGETS),$(eval push.$(TGT): | $(TGT) ; \
        time (set -e && for distro in $(DOCKER_BUILD_VARIANTS); do tag=$(TAG)-$$$${distro}; docker tag $(subst docker.,,$(TGT)):$(DOCKER_IMAGE_TAG) $(HUB)/$(subst docker.,,$(TGT)):$(DOCKER_IMAGE_TAG); docker push $(HUB)/$(subst docker.,,$(TGT)):$(DOCKER_IMAGE_TAG); done)))

