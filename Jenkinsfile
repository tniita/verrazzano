// Copyright (c) 2020, 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

def DOCKER_IMAGE_TAG
def availableRegions = [ "us-ashburn-1", "ap-chuncheon-1", "ap-hyderabad-1", "ap-melbourne-1", "ap-mumbai-1", "ap-osaka-1", "ap-seoul-1", "ap-sydney-1",
                          "ap-tokyo-1", "ca-montreal-1", "ca-toronto-1", "eu-amsterdam-1", "eu-frankfurt-1", "eu-zurich-1", "me-jeddah-1",
                          "sa-saopaulo-1", "uk-london-1", "us-phoenix-1" ]
Collections.shuffle(availableRegions)
def agentLabel = env.JOB_NAME.contains('kind') || env.JOB_NAME.contains('verrazzano-acceptance-test-suite') ? "VM.Standard2.8" : ""

pipeline {
    options {
        skipDefaultCheckout true
    }

    agent {
       docker {
            image "${RUNNER_DOCKER_IMAGE}"
            args "${RUNNER_DOCKER_ARGS}"
            registryUrl "${RUNNER_DOCKER_REGISTRY_URL}"
            registryCredentialsId 'ocir-pull-and-push-account'
            label "${agentLabel}"
        }
    }

    parameters {
        string (name: 'RELEASE_VERSION',
                defaultValue: '',
                description: 'Release version used for the version of helm chart and tag for the image:\n'+
                'When RELEASE_VERSION is not defined, version will be determined by incrementing last minor release version by 1, for example:\n'+
                'When RELEASE_VERSION is v0.1.0, image tag will be v0.1.0 and helm chart version is also v0.1.0.\n'+
                'When RELEASE_VERSION is not specified and last release version is v0.1.0, image tag will be v0.1.1 and helm chart version is also v0.1.1.',
                trim: true)
        string (name: 'RELEASE_DESCRIPTION',
                defaultValue: '',
                description: 'Brief description for the release.',
                trim: true)
        string (name: 'RELEASE_BRANCH',
                defaultValue: 'master',
                description: 'Branch to create release from, change this to enable release from a non master branch, e.g.\n'+
                'When the branch being built is master then release will always be created when RELEASE_BRANCH has the default value - master.\n'+
                'When the branch being built is any non-master branch - release can be created by setting RELEASE_BRANCH to same value as non-master branch, else it is skipped.\n',
                trim: true)
        string (name: 'ACCEPTANCE_TESTS_BRANCH',
                defaultValue: 'master',
                description: 'Branch or tag of verrazzano acceptance tests, on which to kick off the tests',
                trim: true
        )
        choice (description: 'OCI region to launch acceptance tests OKE clusters in', name: 'ACCEPTANCE_TESTS_OKE_REGION',
                // 1st choice is the default value
                choices: availableRegions )
        booleanParam (description: 'Whether to kick off acceptance test run at the end of this build', name: 'RUN_ACCEPTANCE_TESTS', defaultValue: false)
    }

    environment {
        CLUSTER_NAME = 'v8o-kind'
        POST_DUMP_FAILED_FILE = "${WORKSPACE}/post_dump_failed_file.tmp"
        KUBECONFIG = "${WORKSPACE}/test_kubeconfig"
        OCR_CREDS = credentials('ocr-pull-and-push-account')
        OCR_REPO = 'container-registry.oracle.com'
        IMAGE_PULL_SECRET = 'verrazzano-container-registry'
        INSTALL_CONFIG_FILE_KIND = "./tests/e2e/config/scripts/install-verrazzano-kind.yaml"
        INSTALL_PROFILE = "dev"
        VZ_ENVIRONMENT_NAME = "default"

        DOCKER_CI_IMAGE_NAME = 'verrazzano-platform-operator-jenkins'
        DOCKER_PUBLISH_IMAGE_NAME = 'verrazzano-platform-operator'
        DOCKER_IMAGE_NAME = "${env.BRANCH_NAME == 'develop' || env.BRANCH_NAME == 'master' ? env.DOCKER_PUBLISH_IMAGE_NAME : env.DOCKER_CI_IMAGE_NAME}"
        CREATE_LATEST_TAG = "${env.BRANCH_NAME == 'master' ? '1' : '0'}"
        GOPATH = '/home/opc/go'
        GO_REPO_PATH = "${GOPATH}/src/github.com/verrazzano"
        DOCKER_CREDS = credentials('github-packages-credentials-rw')
        DOCKER_EMAIL = credentials('github-packages-email')
        DOCKER_REPO = 'ghcr.io'
        DOCKER_NAMESPACE = 'verrazzano'
        NETRC_FILE = credentials('netrc')
        GITHUB_API_TOKEN = credentials('github-api-token-release-assets')
        GITHUB_RELEASE_USERID = credentials('github-userid-release')
        GITHUB_RELEASE_EMAIL = credentials('github-email-release')
        SERVICE_KEY = credentials('PAGERDUTY_SERVICE_KEY')

    }

    stages {
        stage('Clean workspace and checkout') {
            steps {
                script {
                    checkout scm
                }
                sh """
                    cp -f "${NETRC_FILE}" $HOME/.netrc
                    chmod 600 $HOME/.netrc
                """

                sh """
                    echo "${DOCKER_CREDS_PSW}" | docker login ${env.DOCKER_REPO} -u ${DOCKER_CREDS_USR} --password-stdin
                    rm -rf ${GO_REPO_PATH}/verrazzano
                    mkdir -p ${GO_REPO_PATH}/verrazzano
                    tar cf - . | (cd ${GO_REPO_PATH}/verrazzano/ ; tar xf -)
                """

                script {
                    def props = readProperties file: '.verrazzano-development-version'
                    VERRAZZANO_DEV_VERSION = props['verrazzano-development-version']
                    TIMESTAMP = sh(returnStdout: true, script: "date +%Y%m%d%H%M%S").trim()
                    SHORT_COMMIT_HASH = sh(returnStdout: true, script: "git rev-parse --short HEAD").trim()
                    DOCKER_IMAGE_TAG = "${VERRAZZANO_DEV_VERSION}-${TIMESTAMP}-${SHORT_COMMIT_HASH}"
                }

                println("${params.ACCEPTANCE_TESTS_OKE_REGION}")
            }
        }

        stage('Build') {
            when { not { buildingTag() } }
            steps {
                sh """
                    cd ${GO_REPO_PATH}/verrazzano
                    make docker-push DOCKER_REPO=${env.DOCKER_REPO} DOCKER_NAMESPACE=${env.DOCKER_NAMESPACE} DOCKER_IMAGE_NAME=${DOCKER_IMAGE_NAME} DOCKER_IMAGE_TAG=${DOCKER_IMAGE_TAG} CREATE_LATEST_TAG=${CREATE_LATEST_TAG}
                   """
            }
        }

        stage('Code Checks') {
            when { not { buildingTag() } }
            steps {
                sh """
                    cd ${GO_REPO_PATH}/verrazzano
                    make check
                """
            }
        }

        stage('Third Party License Check') {
            when { not { buildingTag() } }
            steps {
                thirdpartyCheck()
            }
        }

        stage('Copyright Compliance Check') {
            when { not { buildingTag() } }
            steps {
                copyrightScan "${WORKSPACE}"
            }
        }

        stage('Unit Tests') {
            when { not { buildingTag() } }
            steps {
                sh """
                    cd ${GO_REPO_PATH}/verrazzano
                    make -B coverage
                    cp coverage.html ${WORKSPACE}
                    cp coverage.xml ${WORKSPACE}
                    operator/build/scripts/copy-junit-output.sh ${WORKSPACE}
                """
            }
            post {
                always {
                    archiveArtifacts artifacts: '**/coverage.html', allowEmptyArchive: true
                    junit testResults: '**/*test-result.xml', allowEmptyResults: true
                    cobertura(coberturaReportFile: 'coverage.xml',
                      enableNewApi: true,
                      autoUpdateHealth: false,
                      autoUpdateStability: false,
                      failUnstable: true,
                      failUnhealthy: true,
                      failNoReports: true,
                      onlyStable: false,
                      fileCoverageTargets: '100, 0, 0',
                      lineCoverageTargets: '85, 85, 85',
                      packageCoverageTargets: '100, 0, 0',
                    )
                }
            }
        }

        stage('Scan Image') {
            when { not { buildingTag() } }
            steps {
                script {
                    clairScanTemp "${env.DOCKER_REPO}/${env.DOCKER_NAMESPACE}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
                }
            }
            post {
                always {
                    archiveArtifacts artifacts: '**/scanning-report.json', allowEmptyArchive: true
                }
            }
        }

        stage('Generate operator.yaml') {
            when { not { buildingTag() } }
            steps {
                sh """
                    cd ${GO_REPO_PATH}/verrazzano/operator
                    cat config/deploy/verrazzano-platform-operator.yaml | sed -e "s|IMAGE_NAME|${env.DOCKER_REPO}/${env.DOCKER_NAMESPACE}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}|g" > deploy/operator.yaml
                    cat config/crd/bases/install.verrazzano.io_verrazzanos.yaml >> deploy/operator.yaml
                    cat deploy/operator.yaml
                   """
            }
        }

        stage('Integration Tests') {
            when { not { buildingTag() } }
            steps {
                sh """
                    cd ${GO_REPO_PATH}/verrazzano
                    make integ-test DOCKER_REPO=${env.DOCKER_REPO} DOCKER_NAMESPACE=${env.DOCKER_NAMESPACE} DOCKER_IMAGE_NAME=${DOCKER_IMAGE_NAME} DOCKER_IMAGE_TAG=${DOCKER_IMAGE_TAG}
                    operator/build/scripts/copy-junit-output.sh ${WORKSPACE}
                """
            }
            post {
                always {
                    archiveArtifacts artifacts: '**/coverage.html', allowEmptyArchive: true
                    junit testResults: '**/*test-result.xml', allowEmptyResults: true
                }
            }
        }

        stage('install-kind') {
            steps {
                sh """
                    cd ${GO_REPO_PATH}/verrazzano
                    ./tests/e2e/config/scripts/install_kind.sh
                """
            }
        }

        stage("create-image-pull-secrets") {
            steps {
                sh """
                    # Create image pull secret for Verrazzano docker images
                    cd ${GO_REPO_PATH}/verrazzano
                    ./tests/e2e/config/scripts/create-image-pull-secret.sh "${IMAGE_PULL_SECRET}" "${DOCKER_REPO}" "${DOCKER_CREDS_USR}" "${DOCKER_CREDS_PSW}"
                    ./tests/e2e/config/scripts/create-image-pull-secret.sh github-packages "${DOCKER_REPO}" "${DOCKER_CREDS_USR}" "${DOCKER_CREDS_PSW}"
                    ./tests/e2e/config/scripts/create-image-pull-secret.sh ocr "${OCR_REPO}" "${OCR_CREDS_USR}" "${OCR_CREDS_PSW}"
                """
            }
        }

        stage("install-platform-operator") {
            steps {
                sh """
                    cd ${GO_REPO_PATH}/verrazzano

                    # Configure the deployment file to use an image pull secret for branches that have private images
                    if [ "${env.BRANCH_NAME}" == "master" ] || [ "${env.BRANCH_NAME}" == "develop" ]; then
                        echo "Using operator.yaml from Verrazzano repo"
                        cp operator/deploy/operator.yaml /tmp/operator.yaml
                    else
                        echo "Generating operator.yaml based on image name provided: ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
                        ./tests/e2e/config/scripts/process_operator_yaml.sh operator "${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
                    fi

                    # Install the verrazzano-platform-operator
                    cat /tmp/operator.yaml
                    kubectl apply -f /tmp/operator.yaml
                    
                    # make sure ns exists
                    ./tests/e2e/config/scripts/check_verrazzano_ns_exists.sh verrazzano-install
                    
                    # create secret in verrazzano-install ns
                    ./tests/e2e/config/scripts/create-image-pull-secret.sh "${IMAGE_PULL_SECRET}" "${DOCKER_REPO}" "${DOCKER_CREDS_USR}" "${DOCKER_CREDS_PSW}" "verrazzano-install"

                    # Configure the custom resource to install verrazzano on Kind
                    echo "Installing yq"
                    GO111MODULE=on go get github.com/mikefarah/yq/v4
                    export PATH=${HOME}/go/bin:${PATH}
                    ./tests/e2e/config/scripts/process_kind_install_yaml.sh ${INSTALL_CONFIG_FILE_KIND}
                """
            }
        }

        stage("install-verrazzano") {
            steps {
                sh """
                    echo "Waiting for Operator to be ready"
                    cd ${GO_REPO_PATH}/verrazzano
                    kubectl -n verrazzano-install rollout status deployment/verrazzano-platform-operator

                    echo "Installing Verrazzano on Kind"
                    kubectl apply -f ${INSTALL_CONFIG_FILE_KIND}

                    # wait for Verrazzano install to complete
                    ./tests/e2e/config/scripts/wait-for-verrazzano-install.sh
                """
            }
            post {
                always {
                    sh """
                        ## dump out install logs
                        mkdir -p ${WORKSPACE}/verrazzano/operator/scripts/install/build/logs
                        kubectl logs --selector=job-name=verrazzano-install-my-verrazzano > ${WORKSPACE}/verrazzano/operator/scripts/install/build/logs/verrazzano-install.log --tail -1
                        kubectl describe pod --selector=job-name=verrazzano-install-my-verrazzano > ${WORKSPACE}/verrazzano/operator/scripts/install/build/logs/verrazzano-install-job-pod.out
                        echo "Verrazzano Installation logs dumped to verrazzano-install.log"
                        echo "Verrazzano Install pod description dumped to verrazzano-install-job-pod.out"
                        echo "------------------------------------------"
                    """
                }
            }
        }

        stage('acceptance-tests') {
            parallel {
                stage('verify-install') {
                    steps {
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE') {
                            sh """
                                cd ${GO_REPO_PATH}/verrazzano/tests/e2e
                                ginkgo -p --randomizeAllSpecs -v -keepGoing --noColor verify-install/...
                            """
                        }
                    }
                }
            }
        }

        stage('Update operator.yaml') {
            when {
                allOf {
                    not { buildingTag() }
                    anyOf { branch 'master'; branch 'develop' }
                }
            }
            steps {
                sh """
                    cd ${GO_REPO_PATH}/verrazzano/operator
                    git config --global credential.helper "!f() { echo username=\\$DOCKER_CREDS_USR; echo password=\\$DOCKER_CREDS_PSW; }; f"
                    git config --global user.name $DOCKER_CREDS_USR
                    git config --global user.email "${DOCKER_EMAIL}"
                    git checkout -b ${env.BRANCH_NAME}
                    git add deploy/operator.yaml
                    git commit -m "[verrazzano] Update verrazzano-platform-operator image version to ${DOCKER_IMAGE_TAG} in operator.yaml"
                    git push origin ${env.BRANCH_NAME}
                   """
            }
        }
    }

    post {
        always {
            dumpVerrazzanoSystemPods()
            dumpCattleSystemPods()
            dumpNginxIngressControllerLogs()
            dumpVerrazzanoPlatformOperatorLogs()

            archiveArtifacts artifacts: '**/coverage.html,**/logs/**,**/verrazzano_images.txt', allowEmptyArchive: true
            junit testResults: '**/*test-result.xml', allowEmptyResults: true

            sh """
                cd ${GO_REPO_PATH}/verrazzano
                ./tests/e2e/config/scripts/delete-kind-cluster.sh
                if [ -f ${POST_DUMP_FAILED_FILE} ]; then
                  echo "Failures seen during dumping of artifacts, treat post as failed"
                  exit 1
                fi
            """
            deleteDir()
        }
        failure {
            mail to: "${env.BUILD_NOTIFICATION_TO_EMAIL}", from: "${env.BUILD_NOTIFICATION_FROM_EMAIL}",
            subject: "Verrazzano: ${env.JOB_NAME} - Failed",
            body: "Job Failed - \"${env.JOB_NAME}\" build: ${env.BUILD_NUMBER}\n\nView the log at:\n ${env.BUILD_URL}\n\nBlue Ocean:\n${env.RUN_DISPLAY_URL}"
            script {
                if (env.JOB_NAME == "verrazzano/master" || env.JOB_NAME == "verrazzano/develop") {
                    pagerduty(resolve: false, serviceKey: "$SERVICE_KEY", incDescription: "Verrazzano: ${env.JOB_NAME} - Failed", incDetails: "Job Failed - \"${env.JOB_NAME}\" build: ${env.BUILD_NUMBER}\n\nView the log at:\n ${env.BUILD_URL}\n\nBlue Ocean:\n${env.RUN_DISPLAY_URL}")
                    slackSend ( message: "Job Failed - \"${env.JOB_NAME}\" build: ${env.BUILD_NUMBER}\n\nView the log at:\n ${env.BUILD_URL}\n\nBlue Ocean:\n${env.RUN_DISPLAY_URL}" )
                }
            }
        }
    }
}

def dumpVerrazzanoSystemPods() {
    sh """
        cd ${GO_REPO_PATH}/verrazzano/operator
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/operator/scripts/install/build/logs/verrazzano-system-pods.log"
        ./scripts/install/k8s-dump-objects.sh -o pods -n verrazzano-system -m "verrazzano system pods" || echo "failed" > ${POST_DUMP_FAILED_FILE}
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/operator/scripts/install/build/logs/verrazzano-system-certs.log"
        ./scripts/install/k8s-dump-objects.sh -o cert -n verrazzano-system -m "verrazzano system certs" || echo "failed" > ${POST_DUMP_FAILED_FILE}
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/operator/scripts/install/build/logs/verrazzano-system-kibana.log"
        ./scripts/install/k8s-dump-objects.sh -o pods -n verrazzano-system -r "vmi-system-kibana-*" -m "verrazzano system kibana log" -l -c kibana || echo "failed" > ${POST_DUMP_FAILED_FILE}
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/operator/scripts/install/build/logs/verrazzano-system-es-master.log"
        ./scripts/install/k8s-dump-objects.sh -o pods -n verrazzano-system -r "vmi-system-es-master-*" -m "verrazzano system kibana log" -l -c es-master || echo "failed" > ${POST_DUMP_FAILED_FILE}
    """
}

def dumpCattleSystemPods() {
    sh """
        cd ${GO_REPO_PATH}/verrazzano/operator
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/operator/scripts/install/build/logs/cattle-system-pods.log"
        ./scripts/install/k8s-dump-objects.sh -o pods -n cattle-system -m "cattle system pods" || echo "failed" > ${POST_DUMP_FAILED_FILE}
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/operator/scripts/install/build/logs/rancher.log"
        ./scripts/install/k8s-dump-objects.sh -o pods -n cattle-system -r "rancher-*" -m "Rancher logs" -l || echo "failed" > ${POST_DUMP_FAILED_FILE}
    """
}

def dumpNginxIngressControllerLogs() {
    sh """
        cd ${GO_REPO_PATH}/verrazzano/operator
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/operator/scripts/install/build/logs/nginx-ingress-controller.log"
        ./scripts/install/k8s-dump-objects.sh -o pods -n ingress-nginx -r "nginx-ingress-controller-*" -m "Nginx Ingress Controller" -l || echo "failed" > ${POST_DUMP_FAILED_FILE}
    """
}

def dumpVerrazzanoPlatformOperatorLogs() {
    sh """
        ## dump out verrazzano-platform-operator logs
        mkdir -p ${WORKSPACE}/verrazzano-platform-operator/logs
        kubectl -n verrazzano-install logs --selector=app=verrazzano-platform-operator > ${WORKSPACE}/verrazzano-platform-operator/logs/verrazzano-platform-operator-pod.log --tail -1 || echo "failed" > ${POST_DUMP_FAILED_FILE}
        kubectl -n verrazzano-install describe pod --selector=app=verrazzano-platform-operator > ${WORKSPACE}/verrazzano-platform-operator/logs/verrazzano-platform-operator-pod.out || echo "failed" > ${POST_DUMP_FAILED_FILE}
        echo "verrazzano-platform-operator logs dumped to verrazzano-platform-operator-pod.log"
        echo "verrazzano-platform-operator pod description dumped to verrazzano-platform-operator-pod.out"
        echo "------------------------------------------"
    """
}
