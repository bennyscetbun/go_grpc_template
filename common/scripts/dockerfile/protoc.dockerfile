FROM rvolosatovs/protoc

ARG USERID=1000
ARG USERGRP=1000
RUN getent group | grep -E "^[^:]+:[^:]+:${USERGRP}(:.*)?$" > /dev/null || addgroup -g $USERGRP localusergrp
RUN /bin/bash -c "getent passwd | grep -E \"^[^:]+:[^:]+:${USERID}(:.*)?$\" > /dev/null || adduser -u $USERID -G `getent group | grep -E \"^[^:]+:[^:]+:${USERGRP}(:.*)?$\" | cut -f 1 -d :` -H -D localuser"