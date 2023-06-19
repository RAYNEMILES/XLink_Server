#!/usr/bin/env bash
#fixme This script is to stop the service

source ./style_info.cfg
source ./path_info.cfg

#Put config path to the ENV
export CONFIG_NAME=$config_path
day=`date +"%Y-%m-%d"`
echo "==========CONFIG_NAME:${CONFIG_NAME}==========="  >>../logs/openIM.log.${day} 2>&1 &

service_name=${open_im_user_name}
time=`date +"%Y-%m-%d %H:%M:%S"`
echo "=========================================================="  >>../logs/openIM.log.${day} 2>&1 &
echo "=========================================================="  >>../logs/openIM.log.${day} 2>&1 &
echo "=========================================================="  >>../logs/openIM.log.${day} 2>&1 &
echo "==========${service_name} stop time:${time}==========="  >>../logs/openIM.log.${day} 2>&1 &
echo "=========================================================="  >>../logs/openIM.log.${day} 2>&1 &
echo "=========================================================="  >>../logs/openIM.log.${day} 2>&1 &
echo "=========================================================="  >>../logs/openIM.log.${day} 2>&1 &

#Check whether the service exists
name="ps -aux |grep -w $service_name |grep -v grep"
count="${name}| wc -l"
if [ $(eval ${count}) -gt 0 ]; then
    pid="${name}| awk '{print \$2}'"
    echo -e "${SKY_BLUE_PREFIX}Killing service:$service_name pid:$(eval $pid)${COLOR_SUFFIX}"  >>../logs/openIM.log.${day} 2>&1 &
    #kill the service that existed
    kill -9 $(eval $pid)
    echo -e "${SKY_BLUE_PREFIX}service:$service_name was killed ${COLOR_SUFFIX}"  >>../logs/openIM.log.${day} 2>&1 &
fi
