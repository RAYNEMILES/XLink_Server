#!/usr/bin/env bash
#fixme This script is the total startup script
#fixme The full name of the shell script that needs to be started is placed in the need_to_start_server_shell array

source ./path_info.cfg
#Put config path to the ENV
export CONFIG_NAME=$config_path

day=`date +"%Y-%m-%d"`
echo "==========CONFIG_NAME:${CONFIG_NAME}===========">>../logs/openIM.log.${day} 2>&1 &

#fixme Put the shell script name here
need_to_start_server_shell=(
  start_rpc_service.sh
  open_im_push_restart.sh
  open_im_msg_transfer_restart.sh
  open_im_sdk_server_restart.sh
  open_im_msg_gateway_restart.sh
  open_im_demo_restart.sh
  open_im_cron_task_restart.sh
  open_im_game_store_restart.sh
)
time=`date +"%Y-%m-%d %H:%M:%S"`
echo "==========================================================">>../logs/openIM.log.${day} 2>&1 &
echo "==========================================================">>../logs/openIM.log.${day} 2>&1 &
echo "==========================================================">>../logs/openIM.log.${day} 2>&1 &
echo "==========server start time:${time}===========">>../logs/openIM.log.${day} 2>&1 &
echo "==========================================================">>../logs/openIM.log.${day} 2>&1 &
echo "==========================================================">>../logs/openIM.log.${day} 2>&1 &
echo "==========================================================">>../logs/openIM.log.${day} 2>&1 &

for i in ${need_to_start_server_shell[*]}; do
  chmod +x $i
  ./$i
    if [ $? -ne 0 ]; then
        exit -1
  fi
done
