#!/usr/bin/env bash
#Include shell font styles and some basic information
source ./style_info.cfg
source ./path_info.cfg

#Put config path to the ENV
export CONFIG_NAME=$config_path
day=`date +"%Y-%m-%d"`
echo "==========CONFIG_NAME:${CONFIG_NAME}==========="  >>../logs/openIM.log.${day} 2>&1 &

#Check if the service exists
#If it is exists,kill this process
check=`ps aux | grep -w ./${msg_transfer_name} | grep -v grep| wc -l`
if [ $check -ge 1 ]
then
oldPid=`ps aux | grep -w ./${msg_transfer_name} | grep -v grep|awk '{print $2}'`
 kill -9 $oldPid
fi
#Waiting port recycling
sleep 1

cd ${msg_transfer_binary_root}
for ((i = 0; i < ${msg_transfer_service_num}; i++)); do
      nohup ./${msg_transfer_name}    >>../logs/openIM.log.${day} 2>&1 &
done

#Check launched service process
check=`ps aux | grep -w ./${msg_transfer_name} | grep -v grep| wc -l`
if [ $check -ge 1 ]
then
newPid=`ps aux | grep -w ./${msg_transfer_name} | grep -v grep|awk '{print $2}'`
allPorts=""
    echo -e ${SKY_BLUE_PREFIX}"SERVICE START SUCCESS "${COLOR_SUFFIX}
    echo -e ${SKY_BLUE_PREFIX}"SERVICE_NAME: "${COLOR_SUFFIX}${YELLOW_PREFIX}${msg_transfer_name}${COLOR_SUFFIX}
    echo -e ${SKY_BLUE_PREFIX}"PID: "${COLOR_SUFFIX}${YELLOW_PREFIX}${newPid}${COLOR_SUFFIX}
    echo -e ${SKY_BLUE_PREFIX}"LISTENING_PORT: "${COLOR_SUFFIX}${YELLOW_PREFIX}${allPorts}${COLOR_SUFFIX}
else
    echo -e ${YELLOW_PREFIX}${msg_transfer_name}${COLOR_SUFFIX}${RED_PREFIX}"SERVICE START ERROR, PLEASE CHECK openIM.log"${COLOR_SUFFIX}
fi
