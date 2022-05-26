#! /bin/bash

bump_version() {
  currentVersion=$1
  bumpPart=$(printenv BUMP_PART)
  if [ "$bumpPart" = "" ]; then
    bumpPart="patch"
  fi
  #echo "Bump part is '${bumpPart}' version is ${currentVersion}"
  newVersion=$(/tmp/semver bump $bumpPart "${currentVersion}")
  echo "$newVersion"
}

push_new_version() {
  newVersion=$1
  git tag $newVersion

  git push origin $newVersion
}

updateVersion() {
  currentVersion=$1
  newVersion="$(bump_version $currentVersion)"
  echo "New version is ${newVersion}"
  push_new_version "${newVersion}"
}

# ----------------Body--------------------------
ssh-keygen -R bitbucket.org
mkdir -p ~/.ssh/
touch ~/.ssh/known_hosts
echo "ADD bitbucket to known hosts"
ssh-keyscan -t rsa bitbucket.org >> ~/.ssh/known_hosts

# git checkout master
git reset HEAD --hard
# git checkout HEAD --force
# git pull --rebase origin master

if [ ! -f /tmp/semver ]; then
  curl https://raw.githubusercontent.com/fsaintjacques/semver-tool/master/src/semver --output /tmp/semver
  chmod +x /tmp/semver
fi

/tmp/semver --version

defaultVersion=1.0.0
currentTag=$(git describe --tags) || echo "Not found tags"

if [[ "$currentTag" != *"-"* ]]; then
  if [ "$currentTag" = "" ]; then
    echo "Use default version ${defaultVersion}"
    updateVersion $defaultVersion
  else
    echo "Tag has not been changed ${currentTag}"
    currentVersion=$currentTag
  fi
else
  updateVersion $currentTag
fi
