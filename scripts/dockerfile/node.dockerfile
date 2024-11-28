FROM node:23

ARG USERID=1000
ARG USERGRP=1000
RUN groupadd -f -g $USERGRP localusergrp
RUN /bin/bash -c "getent passwd | grep -E \"^[^:]+:[^:]+:${USERID}(:.*)?$\" > /dev/null || useradd -u $USERID -g $USERGRP -m localuser"