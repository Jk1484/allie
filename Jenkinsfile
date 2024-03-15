pipeline {
  environment {
    imagename = "Jk1484/xm"
    dockerImage = ''
  }
  agent any
  stages {
    stage('Cloning Git') {
      steps {
        git([url: 'https://github.com/Jk1484/xm.git', branch: 'master', credentialsId: 'ismailyenigul-github-user-token'])
      }
    }
    stage('Building image') {
      steps{
        script {
          dockerImage = docker.build imagename
        }
      }
    }
  }
}