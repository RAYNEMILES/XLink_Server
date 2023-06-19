#!/usr/bin/env bash

source ./proto_dir.cfg

for ((i = 0; i < ${#all_proto[*]}; i++)); do
  proto=${all_proto[$i]}
  echo "doing proto " $proto ", i = " $i

#  protoc -I ../../../  -I ./ --go_out=plugins=grpc:. $proto
  protoc -I ../../../  -I ./ --go-grpc_out=require_unimplemented_servers=false:. $proto
  protoc -I ../../../  -I ./ --go_out=. $proto
  s=`echo $proto | sed 's/ //g'`
  v=${s//proto/pb.go}
  protoc-go-inject-tag -input=./$v
  echo "protoc --go-grpc_out=require_unimplemented_servers=false:." $proto
  echo "protoc --go_out=." $proto
done
echo "proto file generate success..."

find ./ -type f -path "*.pb.go"|xargs sed -i '' -e 's/\".\/sdk_ws\"/\"Open_IM\/pkg\/proto\/sdk_ws\"/g'
find ./ -type f -path "*.pb.go"|xargs sed -i '' -e 's/\".\/appversion\"/\"Open_IM\/pkg\/proto\/appversion\"/g'
find ./ -type f -path "*.pb.go"|xargs sed -i '' -e 's/\".\/admin_cms\"/\"Open_IM\/pkg\/proto\/admin_cms\"/g'
find ./ -type f -path "*.pb.go"|xargs sed -i '' -e 's/\".\/news\"/\"Open_IM\/pkg\/proto\/news\"/g'
find ./ -type f -path "*.pb.go"|xargs sed -i '' -e 's/\".\/short_video\"/\"Open_IM\/pkg\/proto\/short_video\"/g'
