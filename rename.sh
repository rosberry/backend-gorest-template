#!/bin/bash
echo "Rename project"

read -p 'Project name: ' projectName
echo "Start process"
baseDir=$(pwd)
#mv "$baseDir/src/project" "$baseDir/src/$projectName" | echo "source path renamed"
#grep -rlZ --exclude=$0 '{{project}}' . | xargs -0 sed -i "s/{{project}}/$projectName/g"

go get github.com/novalagung/gorep
cd src/project
$GOPATH/bin/gorep -from "project" -to "$projectName" 
echo "Import renames"

mv "$baseDir/src/project" "$baseDir/src/$projectName" | echo "Source path renamed"

echo "End process"