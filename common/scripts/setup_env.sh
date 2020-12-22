set -e

LOCAL_ARCH=$(uname -m)
export LOCAL_ARCH
# Pass environment set target architecture to build system
if [[ ${TARGET_ARCH} ]]; then
    export TARGET_ARCH
elif [[ ${LOCAL_ARCH} == x86_64 ]]; then
    export TARGET_ARCH=amd64
elif [[ ${LOCAL_ARCH} == armv8* ]]; then
    export TARGET_ARCH=arm64
elif [[ ${LOCAL_ARCH} == aarch64* ]]; then
    export TARGET_ARCH=arm64
elif [[ ${LOCAL_ARCH} == armv* ]]; then
    export TARGET_ARCH=arm
elif [[ ${LOCAL_ARCH} == s390x ]]; then
￼    export TARGET_ARCH=s390x
else
    echo "This system's architecture, ${LOCAL_ARCH}, isn't supported"
    exit 1
fi


LOCAL_OS=$(uname)
export LOCAL_OS
# Pass environment set target operating-system to build system
if [[ ${TARGET_OS} ]]; then
    export TARGET_OS
elif [[ $LOCAL_OS == Linux ]]; then
    export TARGET_OS=linux
    readlink_flags="-f"
elif [[ $LOCAL_OS == Darwin ]]; then
    export TARGET_OS=darwin
    readlink_flags=""
else
    echo "This system's OS, $LOCAL_OS, isn't supported"
    exit 1
fi


TIMEZONE=$(readlink "$readlink_flags" /etc/localtime | sed -e 's/^.*zoneinfo\///')
export TIMEZONE

export TARGET_OUT="${TARGET_OUT:-$(pwd)/out/${TARGET_OS}_${TARGET_ARCH}}"
export TARGET_OUT_LINUX="${TARGET_OUT_LINUX:-$(pwd)/out/linux_${TARGET_ARCH}}"

# Avoid recursive calls to make from attempting to start an additional container
export BUILD_WITH_CONTAINER=0

# For non container build, we need to write env to file
if [[ "${1}" == "envfile" ]]; then
  echo "LOCAL_OS=${LOCAL_OS}"
  echo "TARGET_OS=${TARGET_OS}"
  echo "LOCAL_ARCH=${LOCAL_ARCH}"
  echo "TARGET_ARCH=${TARGET_ARCH}"
  echo "BUILD_WITH_CONTAINER=0"
  echo "TARGET_OUT_LINUX=${TARGET_OUT_LINUX}"
  echo "TARGET_OUT=${TARGET_OUT}"
  echo "TIMEZONE=${TIMEZONE}"
fi

