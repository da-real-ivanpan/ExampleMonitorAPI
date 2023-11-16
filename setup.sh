#!/bin/bash

sudo mkdir /opt/test_service
sudo mkdir /opt/test_service/pub
sudo touch /opt/test_service/test.txt
sudo touch /opt/test_service/myservice.sh
sudo touch /etc/systemd/system/myservice.service
sudo chmod +x /opt/test_service/myservice.sh
