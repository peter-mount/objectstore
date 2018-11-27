// Build properties
properties([
  buildDiscarder(
    logRotator(
      artifactDaysToKeepStr: '',
      artifactNumToKeepStr: '',
      daysToKeepStr: '',
      numToKeepStr: '10'
    )
  ),
  disableConcurrentBuilds(),
  disableResume(),
  pipelineTriggers([
    cron('H H * * *')
  ])
])

// Repository name use, must end with / or be '' for none.
// Setting this to '' will also disable any pushing
repository= 'area51/'

// image prefix - project specific
imagePrefix = 'objectstore'

// The image tag (i.e. repository/image but no version)
imageTag=repository + imagePrefix

// The architectures to build. This is an array of [node,arch]
architectures = [
 ['AMD64', 'amd64'],
 ['ARM64', 'arm64v8']
]

// The image version based on the branch name - master branch is latest in docker
version=BRANCH_NAME
if( version == 'master' ) {
  version = 'latest'
}

// ======================================================================
// Do not modify anything below this point
// ======================================================================

// Build each architecture on each node in parallel
stage( 'Build' ) {
  def builders = [:]
  for( architecture in architectures ) {
    // Need to bind these before the closure, cannot access these as architecture[x]
    def nodeId = architecture[0]
    def arch = architecture[1]
    builders[arch] = {
      node( nodeId ) {
        withCredentials([
          usernameColonPassword(credentialsId: 'artifact-publisher', variable: 'UPLOAD_CRED')]
        ) {
          stage( arch ) {
            checkout scm

            sh './build.sh ' + imageTag + ' ' + arch + ' ' + version

            if( repository != '' ) {
              sh 'docker push ' + imageTag + ':' + arch + '-' + version
            }
          }
        }
      }
    }
  }
  parallel builders
}

// The multiarch build only if we have a repository set
if( repository != '' ) {
  node( 'AMD64' ) {
    stage( "Multiarch Image" ) {
      sh './multiarch.sh' +
        ' ' + imageTag +
        ' ' + version +
        ' ' + architectures.collect { it[1] } .join(' ')
    }
  }
}
