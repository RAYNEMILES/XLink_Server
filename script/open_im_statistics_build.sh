#!/usr/bin/env bash

source ./style_info.cfg
source ./path_info.cfg
source ./function.sh

bin_dir="../bin"
logs_dir="../logs"
sdk_db_dir="../db/sdk/"
#Automatically created when there is no bin, logs folder
if [ ! -d $bin_dir ]; then
  mkdir -p $bin_dir
fi
if [ ! -d $logs_dir ]; then
  mkdir -p $logs_dir
fi
if [ ! -d $sdk_db_dir ]; then
  mkdir -p $sdk_db_dir
fi

service_name=${open_im_statistics_name}
service_path=${open_im_statistics_source_root}
cd $service_path
make install
if [ $? -ne 0 ]; then
  echo -e "${RED_PREFIX}${service_name} build failed ${COLOR_SUFFIX}\n"
  exit -1
  else
    echo -e "${GREEN_PREFIX}${service_name} successfully be built ${COLOR_SUFFIX}\n"
fi
echo -e ${YELLOW_PREFIX}"${service_name} build success"${COLOR_SUFFIX}
