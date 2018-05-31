pipeline {

    agent any

    stages {
        stage('Build app') {
            steps {
                echo 'Building app..'
                sh "rm -rf $GOPATH/src/github.com/sandramarta1912/adserver"
                sh "mkdir -p $GOPATH/src/github.com/sandramarta1912/adserver"
                sh "mv ./*  $GOPATH/src/github.com/sandramarta1912/adserver"
                dir("$GOPATH/src/github.com/sandramarta1912/adserver") {
	        	    sh './build.sh'
                }
            }
        }

        stage('Build containers') {
            steps {
                echo 'Building containers..'
                dir("$GOPATH/src/github.com/sandramarta1912/adserver") {
                    sh "docker build . --no-cache -t adserver:latest"
                }
            }
        }

        stage('Pushing on Docker repo') {
            steps {
                dir("$GOPATH/src/github.com/sandramarta1912/adserver") {
                    sh 'docker login --username martasandra --password oglinda1912'
                    sh 'docker tag adserver:latest martasandra/adserver'
                    sh 'docker push martasandra/adserver'
                }
            }
        }

        stage('Swarm update ... ') {
            environment {
                DOCKER_HOST="tcp://51.15.213.104:2376"
            }
            steps {
                dir("$GOPATH/src/github.com/sandramarta1912/adserver") {
                   sh 'docker service create --name adserver --network my-overlay  --env MYSQL_DSN="root:cms@tcp(mysql:3306)/admin" -p 3001:3001 adserver:latest'
                }
            }
        }
    }
}