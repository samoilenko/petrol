pipeline {
    agent none
    parameters {
        string(name: 'PETROL_STORAGE_PATH', defaultValue: '/storage/petrol.data', description: 'Path to the file')
        string(name: 'DOCKER_SERVER', description: 'Docker server URI + PORT')
    }
    environment {
        PETROL_TELEGRAM_BOT_TOKEN = credentials('PETROL_TELEGRAM_BOT_TOKEN')
        PETROL_TELEGRAM_WEBHOOK_URL = credentials('PETROL_TELEGRAM_WEBHOOK_URL')
    }
    stages {
        stage('Bump version') {
            agent any
            steps {
                sshagent(credentials: ['bitbucket']) {
                    sh "chmod +x ./docker/prod/bump_version.sh"
                    sh './docker/prod/bump_version.sh'
                }
            }
        }
        stage('Deploy') {
            agent any
            steps {
                withDockerServer([credentialsId: 'ProdDocker', uri: "${DOCKER_SERVER}"]) {
                    sh '''
                    #!/bin/bash
                    docker build -f ./docker/prod/Dockerfile --build-arg GIT_HASH=$(git log --pretty=format:'%H' -n 1) --build-arg VERSION=$(git describe --tags --abbrev=0) --build-arg USER=$(id -u -n) --build-arg DATE="$(date)" -t petrol-app .
                    ROOT_API_DIR=${JENKINS_HOME}/jobs/${JOB_NAME}/

                    FILE=$ROOT_API_DIR"petrolVersion.txt"
                    if [ ! -f $FILE ]; then
                       touch $FILE
                       echo 0 > $FILE
                    fi

                    VERSION=`cat $FILE`

                    export COMPOSE_PROJECT_NAME=stack${VERSION}

                    echo "${COMPOSE_PROJECT_NAME}"
                    echo "Turn off old servers"
                    docker-compose -f ./docker/prod/docker-compose.yml down

                    echo "Get new version number"
                    STACK_NUMBER=$(($VERSION%2))

                    export COMPOSE_PROJECT_NAME=stack${STACK_NUMBER}
                    echo "Making containers"
                    docker-compose -f ./docker/prod/docker-compose.yml up --build -d

                    #reset prefix
                    export COMPOSE_PROJECT_NAME=

                    echo $(($VERSION+1)) > $FILE
                    '''
                }
            }
        }
    }
}
