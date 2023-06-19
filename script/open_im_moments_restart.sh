#!/usr/bin/env bash

source ./style_info.cfg
source ./path_info.cfg
source ./function.sh

#Put config path to the ENV
export CONFIG_NAME=$config_path
day=`date +"%Y-%m-%d"`
echo "==========CONFIG_NAME:${CONFIG_NAME}==========="  >>../logs/openIM.log.${day} 2>&1 &

#get the ports for server openImConversationPort from config.yaml
service_port_name=openImConversationPort
list=$(cat $config_path | grep -w $service_port_name | awk -F '[:]' '{print $NF}')
list_to_string $list
api_ports=($ports_array)

#check if the service exists, if it is exists, kill all the processes
check=$(ps aux | grep -w ./${open_im_moments_name} | grep -v grep | wc -l)
while [ $check -ge 1 ]; do
    oldPid=$(ps aux | grep -w ./${open_im_moments_name} | grep -v grep | awk '{print $2}')
    kill -s 9 $oldPid
    echo "kill" ${open_im_moments_name} " oldPid" $oldPid >>../logs/openIM.log 2>&1
    check=$(ps aux | grep -w ./${open_im_moments_name} | grep -v grep | wc -l)
done

#waiting port recycling
sleep 1
cd ${open_im_moments_binary_root}
for ((i = 0; i < ${#api_ports[@]}; i++)); do
    nohup ./${open_im_moments_name} -port ${api_ports[$i]}   >>../logs/openIM.log.${day} 2>&1 &
done

sleep 3
#Check launched service process
check=$(ps aux | grep -w ./${open_im_moments_name} | grep -v grep | wc -l)
if [ $check -ge 1 ]; then
  newPid=$(ps aux | grep -w ./${open_im_moments_name} | grep -v grep | awk '{print $2}')
  ports=$(netstat -netulp | grep -w ${newPid} | awk '{print $4}' | awk -F '[:]' '{print $NF}')
  allPorts=""

  for i in $ports; do
    allPorts=${allPorts}"$i "
  done
  echo -e ${SKY_BLUE_PREFIX}"SERVICE START SUCCESS "${COLOR_SUFFIX}
  echo -e ${SKY_BLUE_PREFIX}"SERVICE_NAME: "${COLOR_SUFFIX}${YELLOW_PREFIX}${open_im_moments_name}${COLOR_SUFFIX}
  echo -e ${SKY_BLUE_PREFIX}"PID: "${COLOR_SUFFIX}${YELLOW_PREFIX}${newPid}${COLOR_SUFFIX}
  echo -e ${SKY_BLUE_PREFIX}"LISTENING_PORT: "${COLOR_SUFFIX}${YELLOW_PREFIX}${allPorts}${COLOR_SUFFIX}
else
  echo -e ${YELLOW_PREFIX}${open_im_moments_name}${COLOR_SUFFIX}${RED_PREFIX}"SERVICE START ERROR, PLEASE CHECK openIM.log"${COLOR_SUFFIX}
fi


