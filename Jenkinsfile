pipeline {
  agent any
  stages {
    stage('Checkout Code, with fix for go modules cache') {
      steps {
        sh '''if [ -d "${GOPATH}" ]; then
  echo "fixing the directory permissions of the Go module chache"
  find "${GOPATH}" -type d -a -not -perm -u=w -ls -exec chmod u+w {} \\;
fi'''
        checkout scm
      }
    }

    stage('Code Checks') {
      parallel {
        stage('go format') {
          steps {
            sh 'make format'
          }
        }

        stage('go lint') {
          steps {
            sh 'make lint'
          }
        }

        stage('go vet') {
          steps {
            sh 'make vet'
          }
        }

      }
    }

    stage('Build') {
      parallel {
        stage('Build') {
          steps {
            sh 'make build'
          }
        }

        stage('Test') {
          steps {
            sh 'make test'
          }
        }

        stage('Coverage') {
          steps {
            sh 'make coverage'
          }
        }

      }
    }

    stage('Package') {
      steps {
        echo 'Package created'
      }
    }

    stage('Acceptance') {
      steps {
        input 'Everything looks good?'
      }
    }

  }
  environment {
    GOPATH = "${WORKSPACE}/.gopath"
    CGO_ENABLED = '0'
  }
  options {
    skipDefaultCheckout(true)
  }
}